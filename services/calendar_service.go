package services

import (
	"log"
	"time"

	"github.com/kajikentaro/meeting-reminder/ui"
	"github.com/kajikentaro/meeting-reminder/utils/xtime"
)

//go:generate mockgen -destination=../mocks/mock_microsoft_repository.go -package=mocks . MicrosoftRepository,UI
type MicrosoftRepository interface {
	FetchCalendarEvents() ([]map[string]interface{}, error)
}

type UI interface {
	ShowMeetingReminder(events []ui.UIEvents)
}

type CalendarService struct {
	repo          MicrosoftRepository
	ui            UI
	watchInterval time.Duration
}

func NewCalendarService(repo MicrosoftRepository, ui UI, watchInterval time.Duration) *CalendarService {
	return &CalendarService{repo: repo, ui: ui, watchInterval: watchInterval}
}

func (s *CalendarService) WaitUntilNextInterval() {
	now := xtime.Now()
	now = now.Truncate(time.Duration(s.watchInterval))
	next := now.Add(s.watchInterval)
	time.Sleep(time.Until(next))
}

func (s *CalendarService) isSameTime(t1, t2 time.Time) bool {
	t1 = t1.Truncate(s.watchInterval)
	t2 = t2.Truncate(s.watchInterval)
	return t1.Equal(t2)
}

func (s *CalendarService) FetchAndDisplayEvents() {
	events, err := s.repo.FetchCalendarEvents()
	if err != nil {
		log.Printf("Error fetching calendar events: %v", err)
		return
	}

	var filteredEvents []ui.UIEvents

	for _, event := range events {
		start, ok := event["start"].(map[string]interface{})
		if !ok {
			log.Printf("Invalid event format: missing 'start' field: %+v", event)
			continue
		}
		startStr, ok := start["dateTime"].(string)
		if !ok {
			log.Printf("Invalid event format: 'start.dateTime' is not a string: %+v", event)
			continue
		}

		startTime, err := parseTime(startStr)
		if err != nil {
			log.Printf("Error parsing start time for event: %+v, error: %v", event, err)
			continue
		}

		subject, ok := event["subject"].(string)
		if !ok {
			log.Printf("Invalid event format: 'subject' is not a string: %+v", event)
			continue
		}

		if !s.isSameTime(startTime, xtime.Now()) {
			continue
		}

		log.Println("Meeting found:", subject, "at", startTime.Format("15:04"))
		location, _ := event["location"].(map[string]interface{})["displayName"].(string)
		filteredEvents = append(filteredEvents, ui.UIEvents{
			Title:     subject,
			StartTime: startTime,
			Link:      location,
		})
	}

	if len(filteredEvents) <= 0 {
		log.Println("No meetings found at this time.")
		return
	}

	s.ui.ShowMeetingReminder(filteredEvents)
}

func (s *CalendarService) StartEventWatcher() {
	log.Println("Starting calendar event watcher...")
	for {
		s.WaitUntilNextInterval()
		s.FetchAndDisplayEvents()
	}
}

func parseTime(s string) (time.Time, error) {
	layout := "2006-01-02T15:04:05.0000000"
	return time.Parse(layout, s)
}
