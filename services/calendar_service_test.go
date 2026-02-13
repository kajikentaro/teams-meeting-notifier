package services

import (
	"sync"
	"testing"
	"time"

	"github.com/kajikentaro/meeting-reminder/mocks"
	"github.com/kajikentaro/meeting-reminder/ui"
	"github.com/kajikentaro/meeting-reminder/utils/xtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func createMockEvent(time time.Time, subject string) map[string]interface{} {
	return map[string]interface{}{
		"start": map[string]interface{}{
			"dateTime": time.Format(TIME_LAYOUT),
		},
		"subject": subject,
		"location": map[string]interface{}{
			"displayName": "Test Location",
		},
	}
}

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
				events = append(events, createMockEvent(event.start, "Test Meeting"))
			}

			repo := mocks.NewMockMicrosoftRepository(ctrl)
			repo.EXPECT().FetchCalendarEvents().Return(events, nil)
			uiMock := mocks.NewMockUI(ctrl)
			expectedEvents := []ui.UIEvents{}
			for _, event := range tc.events {
				if event.shouldFound {
					expectedEvents = append(expectedEvents, ui.UIEvents{
						Title:     "Test Meeting",
						StartTime: event.start,
						Link:      "Test Location",
					})
				}
			}
			if len(expectedEvents) > 0 {
				uiMock.EXPECT().ShowMeetingReminder(expectedEvents).Times(1)
			} else {
				uiMock.EXPECT().ShowMeetingReminder(gomock.Any()).Times(0)
			}

			service := NewCalendarService(repo, uiMock, tc.watchInterval)
			service.FetchAndDisplayEvents()
		})
	}
}

func TestMultipleEventsAtSameTime(t *testing.T) {
	NOW := time.Date(2033, 3, 3, 3, 3, 33, 333, time.UTC)
	xtime.Mock(NOW)
	defer xtime.Unmock()

	eventTime := time.Date(2033, 3, 3, 3, 3, 0, 0, time.UTC)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	events := []map[string]interface{}{
		createMockEvent(eventTime, "Event A"),
		createMockEvent(eventTime, "Event B"),
	}

	repo := mocks.NewMockMicrosoftRepository(ctrl)
	repo.EXPECT().FetchCalendarEvents().Return(events, nil)
	uiMock := mocks.NewMockUI(ctrl)

	uiMock.EXPECT().ShowMeetingReminder([]ui.UIEvents{
		{
			Title:     "Event A",
			StartTime: eventTime,
			Link:      "Test Location",
		},
		{
			Title:     "Event B",
			StartTime: eventTime,
			Link:      "Test Location",
		},
	}).Times(1)

	service := NewCalendarService(repo, uiMock, time.Minute)
	service.FetchAndDisplayEvents()
}

func TestFetchAndDisplayEvents_NoEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockMicrosoftRepository(ctrl)
	uiMock := mocks.NewMockUI(ctrl)

	service := NewCalendarService(repo, uiMock, time.Minute)

	repo.EXPECT().FetchCalendarEvents().Return([]map[string]interface{}{}, nil)
	uiMock.EXPECT().ShowMeetingReminder(gomock.Any()).Times(0)

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
