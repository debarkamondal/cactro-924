package utils

import (
	"fmt"
	"net/http"
	"slices"
)

var allowedOrigins = []string{
	"http://127.0.0.1:8132",
}

func CORS(next http.HandlerFunc) http.HandlerFunc {
	fmt.Println("middleware ran")
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if slices.Contains(allowedOrigins, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Add("Vary", "Origin")
		next(w, r)
	}
}
