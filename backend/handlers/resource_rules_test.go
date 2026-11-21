package handlers

import (
	"reflect"
	"strings"
	"testing"
)

func boolPtr(v bool) *bool { return &v }
func intPtr(v int) *int    { return &v }
func uintPtr(v uint) *uint { return &v }

func TestFlagUpdates(t *testing.T) {
	tests := []struct {
		name    string
		flags   resourceFlags
		columns []string
		want    map[string]interface{}
	}{
		{"garde false et 0", resourceFlags{IsActive: boolPtr(false), Order: intPtr(0)}, []string{"is_active", "order"},
			map[string]interface{}{"is_active": false, "order": 0}},
		{"ignore les champs absents", resourceFlags{IsActive: boolPtr(true)}, []string{"is_active", "is_private", "order"},
			map[string]interface{}{"is_active": true}},
		{"ignore les colonnes non demandées", resourceFlags{IsPrivate: boolPtr(false), OpenInNewTab: boolPtr(false)}, []string{"open_in_new_tab"},
			map[string]interface{}{"open_in_new_tab": false}},
		{"requête vide", resourceFlags{}, []string{"is_active"}, map[string]interface{}{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := flagUpdates(tt.flags, tt.columns...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("flagUpdates = %v, attendu %v", got, tt.want)
			}
		})
	}
}

func TestResolveOwnerGroup(t *testing.T) {
	tests := []struct {
		name      string
		requested *uint
		managed   []uint
		want      uint
		wantOK    bool
	}{
		{"garde le groupe choisi", uintPtr(7), []uint{3, 7}, 7, true},
		{"premier groupe si aucun choix", nil, []uint{3, 7}, 3, true},
		{"premier groupe si choix vide", uintPtr(0), []uint{3, 7}, 3, true},
		{"refuse un groupe non administré", uintPtr(9), []uint{3, 7}, 0, false},
		{"refuse sans groupe administré", nil, nil, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := resolveOwnerGroup(tt.requested, tt.managed)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("resolveOwnerGroup = (%d, %v), attendu (%d, %v)", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestListSearchTooLong(t *testing.T) {
	tests := []struct {
		name   string
		search string
		want   bool
	}{
		{"un caractère", "a", false},
		{"100 lettres arabes (200 octets)", strings.Repeat("ب", 100), false},
		{"101 caractères", strings.Repeat("a", 101), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := listSearchTooLong(tt.search); got != tt.want {
				t.Errorf("listSearchTooLong = %v, attendu %v", got, tt.want)
			}
		})
	}
}

func TestParseBoolFilter(t *testing.T) {
	tests := []struct {
		raw    string
		want   bool
		wantOK bool
	}{
		{"true", true, true},
		{"false", false, true},
		{"", false, false},
		{"oui", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			got, ok := parseBoolFilter(tt.raw)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("parseBoolFilter(%q) = (%v, %v), attendu (%v, %v)", tt.raw, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
