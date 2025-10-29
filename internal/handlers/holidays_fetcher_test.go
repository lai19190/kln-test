package handlers

import (
	"context"
	"encoding/json"
	"kln-test/internal/holidays"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockHolidaysService struct {
}

func (m *mockHolidaysService) GetHolidaysForCountries(ctx context.Context, year int, countryCodes []string) []holidays.CountryResult {
	results := make([]holidays.CountryResult, 0, len(countryCodes))
	for _, code := range countryCodes {
		results = append(results, holidays.CountryResult{
			CountryCode: code,
			Holidays: []holidays.Holiday{
				{Date: "2025-01-01", LocalName: "New Year's Day", Name: "New Year's Day", CountryCode: code},
			},
		})
	}
	return results
}

func TestHolidaysHandler(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{
			name:       "valid request",
			url:        "/public-holidays?year=2025&country=CA&country=DE",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid year",
			url:        "/public-holidays?year=invalid&country=CA",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing year",
			url:        "/public-holidays?country=CA",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing country",
			url:        "/public-holidays?year=2025",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid country code length",
			url:        "/public-holidays?year=2025&country=CAN",
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "year out of range",
			url:        "/public-holidays?year=1000&country=CA",
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	handler := NewHolidaysFetchHandler(&mockHolidaysService{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Expected status code %d, got %d", tt.wantStatus, rec.Code)
			}

			if rec.Code == http.StatusOK {
				var response map[string]interface{}
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if _, ok := response["year"]; !ok {
					t.Error("Response missing year field")
				}
				if _, ok := response["results"]; !ok {
					t.Error("Response missing results field")
				}
			}
		})
	}
}
