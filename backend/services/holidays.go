package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"airboard/models"

	"gorm.io/gorm"
)

// HolidayService gère l'import des jours fériés
type HolidayService struct {
	db *gorm.DB
}

// NewHolidayService crée une nouvelle instance du service
func NewHolidayService(db *gorm.DB) *HolidayService {
	return &HolidayService{db: db}
}

// NagerDateHoliday représente un jour férié retourné par l'API Nager.Date
type NagerDateHoliday struct {
	Date        string   `json:"date"`
	LocalName   string   `json:"localName"`
	Name        string   `json:"name"`
	CountryCode string   `json:"countryCode"`
	Fixed       bool     `json:"fixed"`
	Global      bool     `json:"global"`
	Counties    []string `json:"counties"`
	LaunchYear  *int     `json:"launchYear"`
	Types       []string `json:"types"`
}

// Country représente un pays supporté
type Country struct {
	CountryCode string `json:"countryCode"`
	Name        string `json:"name"`
}

// HolidayImportResult représente le résultat d'un import
type HolidayImportResult struct {
	Imported int      `json:"imported"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors"`
}

// GetAvailableCountries retourne la liste des pays supportés par l'API
func (s *HolidayService) GetAvailableCountries() ([]Country, error) {
	resp, err := http.Get("https://date.nager.at/api/v3/AvailableCountries")
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des pays: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API a retourné le code %d", resp.StatusCode)
	}

	var countries []Country
	if err := json.NewDecoder(resp.Body).Decode(&countries); err != nil {
		return nil, fmt.Errorf("erreur lors du décodage de la réponse: %v", err)
	}

	return countries, nil
}

// FetchHolidays récupère les jours fériés pour un pays et une année donnés
func (s *HolidayService) FetchHolidays(countryCode string, year int) ([]NagerDateHoliday, error) {
	url := fmt.Sprintf("https://date.nager.at/api/v3/PublicHolidays/%d/%s", year, countryCode)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des jours fériés: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("pays non trouvé ou pas de données pour cette année")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API a retourné le code %d", resp.StatusCode)
	}

	var holidays []NagerDateHoliday
	if err := json.NewDecoder(resp.Body).Decode(&holidays); err != nil {
		return nil, fmt.Errorf("erreur lors du décodage de la réponse: %v", err)
	}

	return holidays, nil
}

// ImportHolidays importe les jours fériés d'un pays pour une année donnée
func (s *HolidayService) ImportHolidays(countryCode string, year int, authorID uint, categoryID *uint) (*HolidayImportResult, error) {
	holidays, err := s.FetchHolidays(countryCode, year)
	if err != nil {
		return nil, err
	}

	result := &HolidayImportResult{
		Imported: 0,
		Skipped:  0,
		Errors:   []string{},
	}

	for _, holiday := range holidays {
		// Parser la date
		holidayDate, err := time.Parse("2006-01-02", holiday.Date)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Date invalide pour %s: %v", holiday.Name, err))
			continue
		}

		// Vérifier si ce jour férié existe déjà (même date, même pays, is_holiday=true)
		var existingCount int64
		s.db.Model(&models.Event{}).
			Where("DATE(start_date) = ? AND country_code = ? AND is_holiday = ?", holiday.Date, countryCode, true).
			Count(&existingCount)

		if existingCount > 0 {
			result.Skipped++
			continue
		}

		// Créer l'événement
		event := models.Event{
			Title:       holiday.LocalName,
			Description: fmt.Sprintf(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"%s"}]}]}`, holiday.Name),
			StartDate:   holidayDate,
			IsAllDay:    true,
			IsPublished: true,
			IsHoliday:   true,
			CountryCode: countryCode,
			AuthorID:    authorID,
			CategoryID:  categoryID,
			Color:       "#EF4444", // Rouge pour les jours fériés
			Priority:    "normal",
			Status:      "confirmed",
			Timezone:    "UTC",
		}

		now := time.Now()
		event.PublishedAt = &now

		if err := s.db.Create(&event).Error; err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Erreur lors de la création de %s: %v", holiday.Name, err))
			continue
		}

		result.Imported++
	}

	return result, nil
}

// DeleteHolidaysByCountryAndYear supprime les jours fériés d'un pays pour une année
func (s *HolidayService) DeleteHolidaysByCountryAndYear(countryCode string, year int) (int64, error) {
	startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	result := s.db.Where("is_holiday = ? AND country_code = ? AND start_date >= ? AND start_date <= ?",
		true, countryCode, startOfYear, endOfYear).
		Delete(&models.Event{})

	return result.RowsAffected, result.Error
}
