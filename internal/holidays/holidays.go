package holidays

import (
	"context"
	"sync"
)

// Holiday represents a public holiday
type Holiday struct {
	Date        string   `json:"date"`
	LocalName   string   `json:"localName"`
	Name        string   `json:"name"`
	CountryCode string   `json:"countryCode"`
	Fixed       bool     `json:"fixed"`
	Global      bool     `json:"global"`
	Counties    []string `json:"counties,omitempty"`
	LaunchYear  *int     `json:"launchYear,omitempty"`
	Type        []string `json:"type"`
}

// CountryResult represents the result of fetching holidays for a single country
type CountryResult struct {
	CountryCode string    `json:"countryCode"`
	Holidays    []Holiday `json:"holidays,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// Service handles holiday data retrieval
type Service interface {
	GetHolidaysForCountries(ctx context.Context, year int, countryCodes []string) []CountryResult
}

// HolidayService implements the Service interface
type HolidayService struct {
	client Client
}

// NewService creates a new holiday service
func NewService(client Client) Service {
	return &HolidayService{client: client}
}

// GetHolidaysForCountries fetches holidays for multiple countries concurrently
func (s *HolidayService) GetHolidaysForCountries(ctx context.Context, year int, countryCodes []string) []CountryResult {
	var (
		wg      sync.WaitGroup
		results = make([]CountryResult, len(countryCodes))
	)

	// Process each country concurrently
	for i, countryCode := range countryCodes {
		wg.Add(1)
		go func(index int, code string) {
			defer wg.Done()

			holidays, err := s.client.GetHolidays(ctx, year, code)
			results[index] = CountryResult{
				CountryCode: code,
			}

			if err != nil {
				results[index].Error = err.Error()
			} else {
				results[index].Holidays = holidays
			}
		}(i, countryCode)
	}

	wg.Wait()
	return results
}
