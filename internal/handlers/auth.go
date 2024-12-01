package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"repup/internal/auth"
	"repup/internal/data"
)

func (h *Handlers) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Create OAuth config
	oauthConfig := auth.NewOAuthConfig()

	// Generate random state
	b := make([]byte, 32)
	rand.Read(b)
	state := base64.StdEncoding.EncodeToString(b)

	// Store state in session/cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to Google's consent page
	url := oauthConfig.Config.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// ///////////////////////////////////////////////////////////////
func (h *Handlers) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Get state from cookie
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "State cookie not found")
		return
	}

	// Verify state matches
	if r.URL.Query().Get("state") != stateCookie.Value {
		h.respondWithError(w, http.StatusBadRequest, "State mismatch")
		return
	}

	// Delete the state cookie as it's no longer needed
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// Create OAuth config
	oauthConfig := auth.NewOAuthConfig()

	// Exchange code for token
	token, err := oauthConfig.Config.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Code exchange failed")
		return
	}

	// Get user info from Google
	googleUser, err := oauthConfig.GetGoogleUser(token)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get user info")
		return
	}

	// Convert to our User model
	user := &data.User{
		Email:         googleUser.Email,
		Name:          googleUser.Name,
		OAuthProvider: "google",
		OAuthID:       googleUser.ID,
	}

	// Create or update user in database
	err = h.models.Users.CreateOrUpdate(user)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	jwtToken, err := auth.CreateToken(user)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create token")
		return
	}
	//Set JWT as HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    jwtToken,
		Path:     "/",
		MaxAge:   86400, // 24 hours
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	//  Return success response
	response := map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
		"token":   jwtToken, // Include token in response body for non-browser clients
		"message": "Successfully authenticated",
	}

	h.respondWithJSON(w, http.StatusOK, response)

}
