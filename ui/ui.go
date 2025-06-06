package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kajikentaro/meeting-reminder/utils"
)

type UI struct {
	BrowserPath string
	OutputDir   string
	OpenDir     string
}

var OUTPUT_NAME = "meeting-reminder.html"

func NewUI(browserPath, outputDir, openDir string) *UI {
	if browserPath == "" {
		panic("BROWSER_PATH environment variable is not set")
	}
	if outputDir == "" {
		outputDir = os.TempDir()
	}
	if openDir == "" {
		openDir = outputDir
	}
	return &UI{
		BrowserPath: browserPath,
		OutputDir:   outputDir,
		OpenDir:     openDir,
	}
}

func (u *UI) ShowMeetingReminder(message, link string) {
	html := fmt.Sprintf(`
		<html>
		<head>
			<title>Meeting Reminder</title>
			<style>
				body {
					background: red;
					color: white;
					font-size: 2em;
					display: flex;
					justify-content: center;
					align-items: center;
					height: 100vh;
					margin: 0;
					flex-direction: column;
				}
				a {
					color: yellow;
					font-size: 0.7em;
					margin-top: 2em;
				}
			</style>
		</head>
		<body>
			<div>%s</div>
			<a href="%s">%s</a>
		</body>
		</html>
	`, message, link, link)

	if err := os.MkdirAll(u.OutputDir, 0700); err != nil {
		panic(err)
	}
	filePath := filepath.Join(u.OutputDir, OUTPUT_NAME)
	f, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.WriteString(html)

	url := filepath.Join(u.OpenDir, OUTPUT_NAME)
	utils.ExecCommand(
		u.BrowserPath,
		url,
	)
}

type MeetingReminderDisplayer interface {
	ShowMeetingReminder(message, link string)
}
