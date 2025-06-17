package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kajikentaro/meeting-reminder/auth"
	"github.com/kajikentaro/meeting-reminder/repositories"
	"github.com/kajikentaro/meeting-reminder/services"
	"github.com/kajikentaro/meeting-reminder/ui"
)

// Load environment variables
func loadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func setupLogging() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Error opening log file:", err)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	setupLogging()

	log.Println("Program started")

	// Load environment variables
	loadEnv()

	// Initialize UI
	uiInstance := ui.NewUI(
		os.Getenv("BROWSER_PATH"),
		os.Getenv("OUTPUT_DIR"),
		os.Getenv("OPEN_DIR"),
	)

	// Get configuration information from environment variables
	tenantID := os.Getenv("TENANT_ID")
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	redirectURL := "http://localhost:9091/callback" // Fixed

	// Initialize Auth
	authInstance, err := auth.NewAuth(
		clientID,
		clientSecret,
		redirectURL,
		tenantID,
	)
	if err != nil {
		log.Fatal("Failed to initialize auth:", err)
	}

	// Initialize Repository with Auth
	microsoftRepo := repositories.NewMicrosoftRepository(authInstance)

	// Initialize Calendar Service
	calendarService := services.NewCalendarService(microsoftRepo, uiInstance, time.Minute)

	// Start the event watcher
	calendarService.StartEventWatcher()
}
