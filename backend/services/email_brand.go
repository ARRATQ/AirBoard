package services

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"

	"airboard/models"
)

// EmailBrand : identité visuelle du portail, reprise dans les emails. Les champs sont promus dans
// les données de chaque modèle ({{.Accent}}, {{.Ink}}, {{.PortalURL}}).
type EmailBrand struct {
	Accent    string // couleur de la palette (repères, liserés)
	Ink       string // variante lisible : fond des boutons à texte blanc (contraste ≥ 4,5:1)
	PortalURL string // adresse du portail
}

// Palettes livrées (frontend/src/config/palettes.js, mode clair). ink vide : calculée.
var emailPalettes = map[string]struct{ accent, ink string }{
	"claude":   {"#d97757", "#a8462f"},
	"ambre":    {"#c8853a", ""},
	"indigo":   {"#5c6ef8", "#3d4ed4"},
	"emeraude": {"#10b981", "#047857"},
	"pivoine":  {"#e84f7a", ""},
}

const defaultEmailPalette = "claude"

var hexColor = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// brandFor résout la palette choisie dans Réglages → Apparence, comme stores/app.js.
func brandFor(settings models.AppSettings, publicURL string) EmailBrand {
	accent, ink := paletteAccent(settings)
	if ink == "" {
		ink = inkForLight(accent)
	}
	return EmailBrand{Accent: strings.ToLower(accent), Ink: strings.ToLower(ink), PortalURL: strings.TrimRight(publicURL, "/")}
}

func paletteAccent(settings models.AppSettings) (accent, ink string) {
	slug := settings.ColorPalette
	switch {
	case slug == "custom":
		if hexColor.MatchString(settings.CustomAccentColor) {
			return settings.CustomAccentColor, ""
		}
	case strings.HasPrefix(slug, "cp_"):
		var custom []struct {
			ID     string `json:"id"`
			Accent string `json:"accent"`
		}
		if json.Unmarshal([]byte(settings.CustomPalettes), &custom) == nil {
			for _, p := range custom {
				if p.ID == slug && hexColor.MatchString(p.Accent) {
					return p.Accent, ""
				}
			}
		}
	default:
		if p, ok := emailPalettes[slug]; ok {
			return p.accent, p.ink
		}
	}
	p := emailPalettes[defaultEmailPalette]
	return p.accent, p.ink
}

// --- Couleurs : portage de frontend/src/config/color.js et tokens.js (inkForLight) ---

const minTextContrast = 4.5

func inkForLight(accent string) string {
	if contrastRatio(accent, "#ffffff") >= minTextContrast {
		return accent
	}
	return adjustForContrast(accent, "#ffffff", minTextContrast, true)
}

func hexToRGB(hex string) (r, g, b float64) {
	var ri, gi, bi int
	fmt.Sscanf(hex[1:], "%02x%02x%02x", &ri, &gi, &bi)
	return float64(ri), float64(gi), float64(bi)
}

// clampByte arrondit comme Math.round (moitié vers le haut).
func clampByte(v float64) int {
	return int(math.Max(0, math.Min(255, math.Floor(v+0.5))))
}

func rgbToHex(r, g, b float64) string {
	return fmt.Sprintf("#%02x%02x%02x", clampByte(r), clampByte(g), clampByte(b))
}

func hexToHSL(hex string) (h, s, l float64) {
	if !hexColor.MatchString(hex) {
		return 18, 67, 59
	}
	r255, g255, b255 := hexToRGB(hex)
	r, g, b := r255/255, g255/255, b255/255
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	l = (max + min) / 2
	if max == min {
		return 0, 0, l * 100
	}
	d := max - min
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}
	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	return h / 6 * 360, s * 100, l * 100
}

func hslToHex(h, s, l float64) string {
	hue := math.Mod(math.Mod(h, 360)+360, 360)
	sat := math.Max(0, math.Min(100, s)) / 100
	lig := math.Max(0, math.Min(100, l)) / 100
	c := (1 - math.Abs(2*lig-1)) * sat
	x := c * (1 - math.Abs(math.Mod(hue/60, 2)-1))
	m := lig - c/2
	var r, g, b float64
	switch {
	case hue < 60:
		r, g, b = c, x, 0
	case hue < 120:
		r, g, b = x, c, 0
	case hue < 180:
		r, g, b = 0, c, x
	case hue < 240:
		r, g, b = 0, x, c
	case hue < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return rgbToHex((r+m)*255, (g+m)*255, (b+m)*255)
}

func relativeLuminance(hex string) float64 {
	lin := func(v float64) float64 {
		c := v / 255
		if c <= 0.03928 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	r, g, b := hexToRGB(hex)
	return 0.2126*lin(r) + 0.7152*lin(g) + 0.0722*lin(b)
}

func contrastRatio(a, b string) float64 {
	la, lb := relativeLuminance(a), relativeLuminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// adjustForContrast fait varier la luminosité HSL par pas de 1 % jusqu'au contraste voulu.
func adjustForContrast(hex, against string, ratio float64, darker bool) string {
	h, s, l := hexToHSL(hex)
	step := 1.0
	if darker {
		step = -1
	}
	for lig := l; lig >= 0 && lig <= 100; lig += step {
		candidate := hslToHex(h, s, lig)
		if contrastRatio(candidate, against) >= ratio {
			return candidate
		}
	}
	if darker {
		return "#000000"
	}
	return "#ffffff"
}

// Libellé et couleur de statut d'une annonce (tokens ok / warn / bad / info du portail).
var announcementTypes = map[string]struct{ label, color string }{
	"info":    {"Information", "#1d4ed8"},
	"warning": {"Avertissement", "#b45309"},
	"success": {"Succès", "#15803d"},
	"error":   {"Erreur", "#b91c1c"},
}

func announcementType(t string) (label, color string) {
	if a, ok := announcementTypes[t]; ok {
		return a.label, a.color
	}
	return announcementTypes["info"].label, announcementTypes["info"].color
}
