package utils

type Album struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Artist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TopTracksResponse struct {
	Items []TrackItem `json:"items"`
}
type TrackItem struct {
	DurationMs int    `json:"duration_ms"`
	ID         string `json:"id"`
	Name       string `json:"name"`
	URI        string `json:"uri"`
}
