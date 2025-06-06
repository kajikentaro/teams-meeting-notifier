package services

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MockRepository struct {
	FetchCalendarEventsFunc func() ([]map[string]interface{}, error)
}

func (m *MockRepository) FetchCalendarEvents() ([]map[string]interface{}, error) {
	return m.FetchCalendarEventsFunc()
}

type MockUI struct {
	ShowMeetingReminderFunc func(msg, location string)
	Called                  bool
	Msg                     string
	Location                string
}

func (m *MockUI) ShowMeetingReminder(msg, location string) {
	m.Called = true
	m.Msg = msg
	m.Location = location
	if m.ShowMeetingReminderFunc != nil {
		m.ShowMeetingReminderFunc(msg, location)
	}
}

func TestFetchAndDisplayEvents_MeetingFound(t *testing.T) {
	repo := &MockRepository{
		FetchCalendarEventsFunc: func() ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{
					"start": map[string]interface{}{"dateTime": time.Now().Format("2006-01-02T15:04:05.0000000")},
					"subject": "Test Meeting",
					"location": map[string]interface{}{ "displayName": "Test Room" },
				},
			}, nil
		},
	}
	ui := &MockUI{}
	svc := NewCalendarService(repo, ui, 1)

	svc.FetchAndDisplayEvents()

	assert.True(t, ui.Called)
	assert.Contains(t, ui.Msg, "Test Meeting")
	assert.Equal(t, "Test Room", ui.Location)
}

func TestFetchAndDisplayEvents_NoMeetingFound(t *testing.T) {
	repo := &MockRepository{
		FetchCalendarEventsFunc: func() ([]map[string]interface{}, error) {
			return []map[string]interface{}{}, nil
		},
	}
	ui := &MockUI{}
	svc := NewCalendarService(repo, ui, 1)

	svc.FetchAndDisplayEvents()

	assert.False(t, ui.Called)
}

func TestFetchAndDisplayEvents_ErrorFetching(t *testing.T) {
	repo := &MockRepository{
		FetchCalendarEventsFunc: func() ([]map[string]interface{}, error) {
			return nil, errors.New("fetch error")
		},
	}
	ui := &MockUI{}
	svc := NewCalendarService(repo, ui, 1)

	svc.FetchAndDisplayEvents()

	assert.False(t, ui.Called)
}
