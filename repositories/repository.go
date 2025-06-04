package repositories

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/kajikentaro/meeting-reminder/auth"
)

type Repository struct {
	Auth *auth.Auth
}

func NewRepository(auth *auth.Auth) *Repository {
	return &Repository{Auth: auth}
}

func (r *Repository) FetchCalendarEvents() ([]map[string]interface{}, error) {
	// Fetch access token from the auth struct
	token, err := r.Auth.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	startDateTime := time.Now().Truncate(24 * time.Hour).Format(time.RFC3339)
	endDateTime := time.Now().Truncate(24 * time.Hour).Add(24*time.Hour - time.Second).Format(time.RFC3339)
	graphAPIEndpoint, err := url.Parse("https://graph.microsoft.com/v1.0/me/calendar/calendarView")
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("startDateTime", startDateTime)
	query.Set("endDateTime", endDateTime)
	graphAPIEndpoint.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", graphAPIEndpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	events, ok := result["value"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	var calendarEvents []map[string]interface{}
	for _, event := range events {
		calendarEvents = append(calendarEvents, event.(map[string]interface{}))
	}

	return calendarEvents, nil
}
