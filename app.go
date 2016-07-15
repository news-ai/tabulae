package tabulae

import (
	"encoding/json"
	"net/http"
)

var (
	// The cities that we will serve
	cities = []string{
		"Amsterdam", "San Francisco", "Paris", "New York", "Portland",
	}
)

func init() {
	// Register the index handler to the
	// default DefaultServeMux.
	http.HandleFunc("/", handleIndex)
}

func handleIndex(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(rw)
	encoder.Encode(cities)
}
