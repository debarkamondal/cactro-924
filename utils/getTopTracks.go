package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func GetTopTracks(userToken string, r *http.Request) (TopTracksResponse, error) {

	queryParams := r.URL.Query()
	timeRange := queryParams.Get("time_range")
	if timeRange == "" {
		timeRange = "long_term"
	}
	limit := queryParams.Get("limit")
	if limit == "" {
		limit = "10"
	}

	// Construct the URL for top tracks
	topTracksURL, err := url.Parse("https://api.spotify.com/v1/me/top/tracks")
	if err != nil {
		return TopTracksResponse{}, err
	}
	topTracksQueryParams := topTracksURL.Query()
	topTracksQueryParams.Set("time_range", timeRange)
	topTracksQueryParams.Set("limit", limit)
	topTracksURL.RawQuery = topTracksQueryParams.Encode()

	req, err := http.NewRequest("GET", topTracksURL.String(), nil)
	if err != nil {
		return TopTracksResponse{}, err
	}
	req.Header.Set("Authorization", "Bearer "+userToken)

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return TopTracksResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Printf("Spotify /me/top/tracks error: %v, body: %s", resp.Status, string(bodyBytes))
		if resp.StatusCode == http.StatusUnauthorized {
			return TopTracksResponse{}, err
		} else {
			return TopTracksResponse{}, err
		}
	}

	var topTracks TopTracksResponse
	if err := json.NewDecoder(resp.Body).Decode(&topTracks); err != nil {
		return TopTracksResponse{}, err
	}
	return topTracks, nil

}
