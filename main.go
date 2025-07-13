package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url" // For encoding form data
	"os"

	"github.com/debarkamondal/cactro-924/spotify"
	"github.com/debarkamondal/cactro-924/utils"
)

var accessToken string

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /spotify/login", utils.CORS(spotify.Login))
	mux.HandleFunc("GET /spotify/play", utils.CORS(spotify.Play))
	mux.HandleFunc("GET /callback/spotify", spotify.CallbackHandler)
	mux.HandleFunc("GET /spotify", func(w http.ResponseWriter, r *http.Request) {

		formData := url.Values{}
		formData.Set("grant_type", "client_credentials")
		formData.Set("client_id", os.Getenv("SPOTIFY_CLIENT_ID"))
		formData.Set("client_secret", os.Getenv("SPOTIFY_CLIENT_SECRET"))

		// Create a new POST request
		// The body needs to be an io.Reader, so we convert the form data string to a bytes.Reader
		req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", bytes.NewBufferString(formData.Encode()))
		if err != nil {
			log.Fatalf("Error creating request: %v", err)
		}

		// Set the Content-Type header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// Create an HTTP client
		client := &http.Client{}

		// Perform the request
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("Error performing request: %v", err)
		}
		defer resp.Body.Close() // Ensure the response body is closed

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Error reading response body: %v", err)
		}

		var authResponse map[string]string
		json.Unmarshal(body, &authResponse)
		accessToken = authResponse["access_token"]

		req, err = http.NewRequest("GET", "https://api.spotify.com/v1/me", nil) // For GET, the body is typically nil
		if err != nil {
			log.Fatalf("Error creating request: %v", err)
		}

		// Set the Authorization header
		req.Header.Set("Authorization", "Bearer "+accessToken)

		// Perform the request
		resp, err = client.Do(req)
		if err != nil {
			log.Fatalf("Error performing request: %v", err)
		}
		defer resp.Body.Close() // Ensure the response body is closed

		// Read the response body
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Error reading response body: %v", err)
		}

		var data map[string]any
		json.Unmarshal(body, &data)
		fmt.Printf("Parsed Data: %+v\n", data)
	})
	fmt.Println("Listening on port 8132")
	if err := http.ListenAndServe(":8132", mux); err != nil {
		fmt.Println(err)
		fmt.Println("Couldn't initiate server on port 8132")
	}
}
