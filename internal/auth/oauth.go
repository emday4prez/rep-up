package auth

import (
	"encoding/json"
	"os"
	"repup/internal/data"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleUser represents the data we get back from Google
type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// OAuthConfig wraps our OAuth2 config
type OAuthConfig struct {
	Config *oauth2.Config
}

// NewOAuthConfig creates a new OAuth configuration
func NewOAuthConfig() *OAuthConfig {
	return &OAuthConfig{
		Config: &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

// GetGoogleUser gets user info from Google's API
func (c *OAuthConfig) GetGoogleUser(token *oauth2.Token) (*GoogleUser, error) {
	client := c.Config.Client(oauth2.NoContext, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var googleUser GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, err
	}

	return &googleUser, nil
}

// ToUser converts a GoogleUser to our internal User model
func (gu *GoogleUser) ToUser() *data.User {
	return &data.User{
		Email:         gu.Email,
		Name:          gu.Name,
		OAuthProvider: "google",
		OAuthID:       gu.ID,
	}
}
