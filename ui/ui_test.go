package ui

import (
	"testing"
	"time"
)

func TestUI(t *testing.T) {
	ui := NewUI("echo", "./", "./")
	events := []UIEvents{
		{
			Title:     "[Sample Sample] Sample Sample Title",
			StartTime: time.Now(),
			Link:      "https://example.com/sample-link",
		},
		{
			Title:     "[Sample Sample] Sample Sample Title aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			StartTime: time.Now(),
			Link:      "https://example.com/sample-link",
		},
		{
			Title:     "[Sample Sample] Sample Sample Only Title",
			StartTime: time.Now(),
		},
	}
	ui.ShowMeetingReminder(events)
}
