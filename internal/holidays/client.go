package holidays

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client interface for holiday API calls
type Client interface {
	GetHolidays(ctx context.Context, year int, countryCode string) ([]Holiday, error)
}

// HTTPClient implements the Client interface using the Nager.Date API
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new holiday API client
func NewClient() *HTTPClient {
	return &HTTPClient{
		baseURL: "https://date.nager.at/api/v3",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetHolidays fetches holidays for a specific country and year
func (c *HTTPClient) GetHolidays(ctx context.Context, year int, countryCode string) ([]Holiday, error) {
	url := fmt.Sprintf("%s/PublicHolidays/%d/%s", c.baseURL, year, countryCode)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch holidays: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var holidays []Holiday
	if err := json.NewDecoder(resp.Body).Decode(&holidays); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return holidays, nil
}
