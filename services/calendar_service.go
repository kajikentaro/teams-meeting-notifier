package services

import (
	"fmt"
	"log"
	"time"

	"github.com/kajikentaro/meeting-reminder/repositories"
	"github.com/kajikentaro/meeting-reminder/ui"
)

type CalendarService struct {
	Repo                 repositories.CalendarEventFetcher
	UI                   ui.MeetingReminderDisplayer
	WatchIntervalMinutes int
}

func NewCalendarService(repo repositories.CalendarEventFetcher, ui ui.MeetingReminderDisplayer, watchIntervalMinutes int) *CalendarService {
	return &CalendarService{Repo: repo, UI: ui, WatchIntervalMinutes: watchIntervalMinutes}
}

func (s *CalendarService) WaitUntilNextInterval() {
	now := time.Now()
	minutes := now.Minute()
	seconds := now.Second()
	nanoseconds := now.Nanosecond()
	waitMin := s.WatchIntervalMinutes - (minutes % s.WatchIntervalMinutes)
	if waitMin == s.WatchIntervalMinutes && (seconds > 0 || nanoseconds > 0) {
		waitMin = s.WatchIntervalMinutes
	}
	wait := time.Duration(waitMin)*time.Minute - time.Duration(seconds)*time.Second - time.Duration(nanoseconds)
	time.Sleep(wait)
}

func isSameTimeMinute(t1, t2 time.Time) bool {
	t1 = t1.Truncate(time.Minute)
	t2 = t2.Truncate(time.Minute)
	return t1.Equal(t2)
}

func (s *CalendarService) FetchAndDisplayEvents() {
	events, err := s.Repo.FetchCalendarEvents()
	if err != nil {
		log.Printf("Error fetching calendar events: %v", err)
		return
	}

	found := false

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

		if !isSameTimeMinute(startTime, time.Now()) {
			continue
		}

		found = true
		log.Println("Meeting found:", subject, "at", startTime.Format("15:04"))
		location, _ := event["location"].(map[string]interface{})["displayName"].(string)
		msg := fmt.Sprintf("%s<br/><br/>Start: %s", subject, startTime.Format("15:04"))
		s.UI.ShowMeetingReminder(msg, location)
	}

	if !found {
		log.Println("No meetings found at this time.")
	}
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
