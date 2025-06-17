package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kajikentaro/meeting-reminder/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

type Auth struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	TenantID     string
	OAuth2Config *oauth2.Config
	Token        *oauth2.Token
}

func NewAuth(clientID, clientSecret, redirectURL, tenantID string) (*Auth, error) {
	endpoint := microsoft.AzureADEndpoint(tenantID)
	authInstance := &Auth{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		TenantID:     tenantID,
		OAuth2Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"https://graph.microsoft.com/Calendars.Read", "offline_access"},
			Endpoint:     endpoint,
		},
	}

	// Check for saved token
	token, err := loadToken()
	if err == nil {
		log.Println("Loaded saved token, checking validity...")
		authInstance.Token = token
		if _, err = authInstance.GetAccessToken(); err == nil {
			log.Println("Saved token is valid, using it for authentication.")
			return authInstance, nil
		}
		log.Println("Saved token is invalid, starting authentication process...")
	} else {
		log.Println("No saved token found, starting authentication process...")
	}

	token, err = authenticate(authInstance.OAuth2Config)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}
	authInstance.Token = token

	if err := saveToken(authInstance.Token); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	return authInstance, nil
}

func (a *Auth) GetAccessToken() (*oauth2.Token, error) {
	if a.Token.Valid() {
		return a.Token, nil
	}

	if a.Token.RefreshToken != "" {
		log.Println("Refreshing access token...")
		ts := a.OAuth2Config.TokenSource(context.Background(), a.Token)
		token, err := ts.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
		a.Token = token
		if err := saveToken(a.Token); err != nil {
			log.Printf("Failed to save refreshed token: %v", err)
		}
		return a.Token, nil
	}

	return nil, fmt.Errorf("no valid refresh token available")
}

func authenticate(config *oauth2.Config) (*oauth2.Token, error) {
	state := "random_state" // Random string for CSRF protection
	authURL := config.AuthCodeURL(state)

	log.Printf("Open the following URL in your browser to authenticate:\n%s\n", authURL)

	codeCh := make(chan string)
	srv := &http.Server{Addr: ":9091"}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Code not found", http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "Authentication completed. You can close this window.")
		codeCh <- code
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	_ = openBrowser(authURL)

	code := <-codeCh

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)

	token, err := config.Exchange(context.TODO(), code)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func saveToken(token *oauth2.Token) error {
	tokenPath, err := GetTokenFilePath()
	if err != nil {
		return err
	}
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return os.WriteFile(tokenPath, data, 0600)
}

func loadToken() (*oauth2.Token, error) {
	tokenPath, err := GetTokenFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, err
	}
	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

func GetTokenFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	configDir = filepath.Join(configDir, "meeting-reminder")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", err
	}
	tokenPath := filepath.Join(configDir, "token.json")
	return tokenPath, nil
}

func openBrowser(url string) error {
	// Branch commands by OS
	switch runtime.GOOS {
	case "windows":
		return utils.ExecCommand("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		return utils.ExecCommand("open", url)
	case "linux":
		return utils.ExecCommand("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}
}
