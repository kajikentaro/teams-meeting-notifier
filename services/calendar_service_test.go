package services

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kajikentaro/meeting-reminder/mocks"
	"github.com/kajikentaro/meeting-reminder/utils/xtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var TIME_LAYOUT = "2006-01-02T15:04:05.0000000"

func TestFetchAndDisplayEvents(t *testing.T) {
	NOW := time.Date(2033, 3, 3, 3, 3, 33, 333, time.UTC)
	xtime.Mock(NOW)
	defer xtime.Unmock()

	type Event struct {
		start       time.Time
		shouldFound bool
	}
	type TestCase struct {
		title         string
		watchInterval time.Duration
		events        []Event
	}

	testCases := []TestCase{
		{
			title:         "Only expected events should be found (Interval: 1 minute)",
			watchInterval: time.Minute,
			events: []Event{
				{start: time.Date(2033, 3, 3, 3, 2, 59, 0, time.UTC), shouldFound: false},
				{start: time.Date(2033, 3, 3, 3, 3, 0, 0, time.UTC), shouldFound: true},
				{start: time.Date(2033, 3, 3, 3, 3, 59, 0, time.UTC), shouldFound: true},
				{start: time.Date(2033, 3, 3, 3, 4, 0, 0, time.UTC), shouldFound: false},
			},
		},
		{
			title:         "Only expected events should be found (Interval: 5 minutes)",
			watchInterval: 5 * time.Minute,
			events: []Event{
				{start: time.Date(2033, 3, 3, 2, 59, 0, 0, time.UTC), shouldFound: false},
				{start: time.Date(2033, 3, 3, 3, 0, 0, 0, time.UTC), shouldFound: true},
				{start: time.Date(2033, 3, 3, 3, 4, 59, 0, time.UTC), shouldFound: true},
				{start: time.Date(2033, 3, 3, 3, 5, 0, 0, time.UTC), shouldFound: false},
			},
		},
		{
			title:         "Only expected events should be found (Interval: 1 hours)",
			watchInterval: time.Hour,
			events: []Event{
				{start: time.Date(2033, 3, 3, 2, 59, 0, 0, time.UTC), shouldFound: false},
				{start: time.Date(2033, 3, 3, 3, 0, 0, 0, time.UTC), shouldFound: true},
				{start: time.Date(2033, 3, 3, 3, 59, 0, 0, time.UTC), shouldFound: true},
				{start: time.Date(2033, 3, 3, 4, 0, 0, 0, time.UTC), shouldFound: false},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			events := []map[string]interface{}{}

			for _, event := range tc.events {
				events = append(events, map[string]interface{}{
					"start": map[string]interface{}{
						"dateTime": event.start.Format(TIME_LAYOUT),
					},
					"subject": "Test Meeting",
					"location": map[string]interface{}{
						"displayName": "Test Location",
					},
				})
			}

			repo := mocks.NewMockMicrosoftRepository(ctrl)
			repo.EXPECT().FetchCalendarEvents().Return(events, nil)
			ui := mocks.NewMockUI(ctrl)
			for _, event := range tc.events {
				if event.shouldFound {
					ui.EXPECT().ShowMeetingReminder(
						fmt.Sprintf("Test Meeting<br/><br/>Start: %s", event.start.Format("15:04")),
						"Test Location",
					).Times(1)
				}
			}

			service := NewCalendarService(repo, ui, tc.watchInterval)
			service.FetchAndDisplayEvents()
		})
	}
}

func TestFetchAndDisplayEvents_NoEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockMicrosoftRepository(ctrl)
	ui := mocks.NewMockUI(ctrl)

	service := NewCalendarService(repo, ui, time.Minute)

	repo.EXPECT().FetchCalendarEvents().Return([]map[string]interface{}{}, nil)
	ui.EXPECT().ShowMeetingReminder(gomock.Any(), gomock.Any()).Times(0)

	service.FetchAndDisplayEvents()
}

func TestIsSameTime(t *testing.T) {
	service := NewCalendarService(nil, nil, time.Minute)

	t1 := time.Date(2025, 6, 10, 15, 4, 0, 0, time.UTC)
	t2 := time.Date(2025, 6, 10, 15, 4, 30, 0, time.UTC)
	t3 := time.Date(2025, 6, 10, 15, 5, 0, 0, time.UTC)

	assert.True(t, service.isSameTime(t1, t2))
	assert.False(t, service.isSameTime(t1, t3))
}

func TestWaitUntilNextInterval(t *testing.T) {
	service := NewCalendarService(nil, nil, time.Second)

	isFinished := false
	var lock sync.Mutex

	go func() {
		service.WaitUntilNextInterval()

		lock.Lock()
		isFinished = true
		lock.Unlock()
	}()

	lock.Lock()
	require.False(t, isFinished)
	lock.Unlock()

	time.Sleep(1100 * time.Millisecond) // Wait slightly more than 1 second

	lock.Lock()
	require.True(t, isFinished)
	lock.Unlock()
}
