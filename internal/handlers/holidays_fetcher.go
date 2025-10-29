package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"kln-test/internal/holidays"

	"github.com/go-playground/validator/v10"
)

// HolidaysRequest represents the query parameters for the holidays endpoint
type HolidaysRequest struct {
	Year      int      `validate:"required,min=1900,max=2100"`
	Countries []string `validate:"required,min=1,dive,required,len=2"`
}

// HolidaysFetchHandler handles public holiday requests
type HolidaysFetchHandler struct {
	validator *validator.Validate
	service   holidays.Service
}

type HolidaysResponse struct {
	Year    int                      `json:"year"`
	Results []holidays.CountryResult `json:"results"`
}

// NewHolidaysFetchHandler creates a new holidays handler
func NewHolidaysFetchHandler(service holidays.Service) *HolidaysFetchHandler {
	return &HolidaysFetchHandler{
		validator: validator.New(),
		service:   service,
	}
}

// ServeHTTP handles HTTP requests for public holidays
func (h *HolidaysFetchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	yearStr := r.URL.Query().Get("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, "Invalid year parameter", http.StatusBadRequest)
		return
	}

	countries := r.URL.Query()["country"]
	if len(countries) == 0 {
		http.Error(w, "At least one country parameter is required", http.StatusBadRequest)
		return
	}

	// Validate request
	req := HolidaysRequest{
		Year:      year,
		Countries: countries,
	}

	if err := h.validator.Struct(req); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	// Get holidays for all countries concurrently
	results := h.service.GetHolidaysForCountries(r.Context(), year, countries)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HolidaysResponse{
		Year:    year,
		Results: results,
	})
}
