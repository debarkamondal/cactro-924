package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /spotify", func(w http.ResponseWriter, r *http.Request) {
		body:= &map[string]any{
			"test":"hello",
		}
		json.NewEncoder(w).Encode(body)
	})
	fmt.Println("Listening on port 8132")
	if err := http.ListenAndServe(":8132", mux); err != nil {
		fmt.Println(err)
		fmt.Println("Couldn't initiate server on port 8132")
	}
}

