package spotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/debarkamondal/cactro-924/utils"
	"github.com/google/uuid"
)

const (
	spotifyAuthURL  = "https://accounts.spotify.com/authorize"
	spotifyTokenURL = "https://accounts.spotify.com/api/token"
	spotifyMeURL    = "https://api.spotify.com/v1/me"

	spotifyScopes = "user-modify-playback-state user-read-private user-read-email user-top-read"
)

var accessToken string

func Play(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	context_uri := queries.Get("context_uri")
	type PlayOffset struct {
		Position int    `json:"position,omitempty"` // 0-indexed position of the item in the context
		URI      string `json:"uri,omitempty"`      // URI of the item to start at
	}

	// PlayOptions represents the JSON body for the Start/Resume Playback request.
	type PlayOptions struct {
		ContextURI string      `json:"context_uri,omitempty"` // spotify:album:xxx, spotify:playlist:xxx, etc.
		URIs       []string    `json:"uris,omitempty"`        // Array of track URIs (if playing specific tracks)
		Offset     *PlayOffset `json:"offset,omitempty"`      // Optional: Start playback at a specific offset
		PositionMs int         `json:"position_ms,omitempty"` // Optional: Start playback at a specific position in milliseconds
		DeviceID   string      `json:"device_id,omitempty"`   // Optional: Device to play on (if omitted, uses active device)
		Play       bool        `json:"play,omitempty"`        // Optional: true or false (Spotify usually assumes true for this endpoint)
	}
	accessToken, err := r.Cookie("access_token")
	if err != nil {
		http.Error(w, "Token not found", http.StatusBadRequest)
		return
	}

	// Build the request URL. You can optionally add a device_id as a query parameter
	// For simplicity, we'll assume the device_id is part of the PlayOptions body for now.
	// If you want to use the query parameter:
	// playURL := spotifyPlayURL
	// if options.DeviceID != "" {
	// 	playURL = fmt.Sprintf("%s?device_id=%s", spotifyPlayURL, options.DeviceID)
	// 	options.DeviceID = "" // Clear it from body if sending as query param
	// }

	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(map[any]any{
		"context_uri": context_uri,
	})
	req, err := http.NewRequest("PUT", "https://api.spotify.com/v1/me/player/play", bytes.NewBuffer([]byte{}))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken.Value)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Spotify's Playback endpoint usually returns 204 No Content for success
	if resp.StatusCode != http.StatusNoContent {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func Login(w http.ResponseWriter, r *http.Request) {

	// Not addding this to db for CSRF protection for evaluation
	state := uuid.New().String()

	authURL, err := url.Parse(spotifyAuthURL)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Error parsing Spotify Auth URL: %v", err)
		return
	}

	queryParams := authURL.Query()
	queryParams.Set("response_type", "code")
	queryParams.Set("client_id", os.Getenv("SPOTIFY_CLIENT_ID"))
	queryParams.Set("scope", spotifyScopes)
	queryParams.Set("redirect_uri", os.Getenv("SPOTIFY_REDIRECT_URI"))
	queryParams.Set("state", state)
	authURL.RawQuery = queryParams.Encode()

	http.Redirect(w, r, authURL.String(), http.StatusFound)
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	code := query.Get("code")
	errorParam := query.Get("error")

	if errorParam != "" {
		http.Error(w, fmt.Sprintf("Spotify authorization error: %s", errorParam), http.StatusUnauthorized)
		return
	}

	if code == "" {
		http.Error(w, "Authorization code not found", http.StatusBadRequest)
		return
	}

	// Verify the state to prevent CSRF.

	// Exchange the authorization code for an access token
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", code)
	formData.Set("redirect_uri", os.Getenv("SPOTIFY_REDIRECT_URI"))
	formData.Set("client_id", os.Getenv("SPOTIFY_CLIENT_ID"))
	formData.Set("client_secret", os.Getenv("SPOTIFY_CLIENT_SECRET"))

	req, err := http.NewRequest("POST", spotifyTokenURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Error creating token request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		log.Printf("Error performing token exchange request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Spotify token exchange error: %v, body: %s", resp.Status, string(bodyBytes))
		http.Error(w, "Failed to get access token from Spotify", http.StatusInternalServerError)
		return
	}

	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Error decoding token response: %v", err)
		return
	}

	userToken := tokenResponse.AccessToken
	topTracks, err := utils.GetTopTracks(userToken, r)

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Domain:   "",
		Path:     "/",
		Value:    userToken,
		MaxAge:   10800,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(topTracks.Items)

}
