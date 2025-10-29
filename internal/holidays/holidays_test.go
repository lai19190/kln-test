package holidays

import (
	"context"
	"testing"
)

type mockClient struct {
	holidays map[string][]Holiday
	err      error
}

func (m *mockClient) GetHolidays(ctx context.Context, year int, countryCode string) ([]Holiday, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.holidays[countryCode], nil
}

func TestGetHolidaysForCountries(t *testing.T) {
	testHolidays := map[string][]Holiday{
		"CA": {
			{Date: "2025-01-01", LocalName: "New Year's Day", Name: "New Year's Day", CountryCode: "CA"},
		},
		"DE": {
			{Date: "2025-01-01", LocalName: "Neujahr", Name: "New Year's Day", CountryCode: "DE"},
		},
	}

	mockClient := &mockClient{holidays: testHolidays}
	service := NewService(mockClient)

	results := service.GetHolidaysForCountries(context.Background(), 2025, []string{"CA", "DE"})

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	for _, result := range results {
		holidays, ok := testHolidays[result.CountryCode]
		if !ok {
			t.Errorf("Unexpected country code: %s", result.CountryCode)
			continue
		}

		if len(result.Holidays) != len(holidays) {
			t.Errorf("Expected %d holidays for %s, got %d", len(holidays), result.CountryCode, len(result.Holidays))
		}
	}
}
