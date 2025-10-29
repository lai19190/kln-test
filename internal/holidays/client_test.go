package holidays

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPClient(t *testing.T) {
	testHolidays := []Holiday{
		{Date: "2025-01-01", LocalName: "New Year's Day", Name: "New Year's Day", CountryCode: "CA"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/PublicHolidays/2025/CA" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(testHolidays)
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL

	holidays, err := client.GetHolidays(context.Background(), 2025, "CA")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(holidays) != len(testHolidays) {
		t.Errorf("Expected %d holidays, got %d", len(testHolidays), len(holidays))
	}
}
