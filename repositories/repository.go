package repositories

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/kajikentaro/meeting-reminder/auth"
	"github.com/kajikentaro/meeting-reminder/utils/xtime"
)

type MicrosoftRepository struct {
	Auth *auth.Auth
}

func NewMicrosoftRepository(auth *auth.Auth) *MicrosoftRepository {
	return &MicrosoftRepository{Auth: auth}
}

func (r *MicrosoftRepository) FetchCalendarEvents() ([]map[string]interface{}, error) {
	// Fetch access token from the auth struct
	token, err := r.Auth.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// DOC: https://learn.microsoft.com/en-us/graph/api/user-list-calendarview
	graphAPIEndpoint, err := url.Parse("https://graph.microsoft.com/v1.0/me/calendar/calendarView")
	if err != nil {
		return nil, err
	}

	startDateTime := xtime.Now().Truncate(24 * time.Hour).Format(time.RFC3339)
	endDateTime := xtime.Now().Truncate(24 * time.Hour).Add(24*time.Hour - time.Second).Format(time.RFC3339)
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
