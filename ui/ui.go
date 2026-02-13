package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

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

type UIEvents struct {
	Title     string
	StartTime time.Time
	Link      string
}

func (u *UI) ShowMeetingReminder(events []UIEvents) {
	html := `<html>
	<head>
		<title>Meeting Reminder</title>
		<style>
			body {
				background: #d70036ff;
				color: white;
				display: flex;
				justify-content: center;
				align-items: center;
				margin: 0;
				flex-direction: column;
				min-height: 100%;
				gap: 20px;
				font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
			}
			a {
				color: yellow;
			}
			div.event {
				background: #0078D7;
				border: 2px solid white;
				padding: 1rem 3rem;
				border-radius: 10px;
			}
		</style>
	</head>
	<body>
		<h1>Meeting is starting now!</h1>`

	for _, event := range events {
		timeStr := event.StartTime.Format("15:04")

		html += fmt.Sprintf(`
			<div class="event">
				<h2>%s</h2>
				<h3>Start Time: %s</h3>
				<a href="%s">%s</a>
			</div>
		`, event.Title, timeStr, event.Link, event.Link)
	}

	html += `
	</body>
</html>`

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
