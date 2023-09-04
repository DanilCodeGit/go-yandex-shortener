package main

import (
	"fmt"
	"io"
	"net/http"
)

var urlStore = make(map[string]string)

func ShortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the URL from the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Convert the request body to a string
	url := string(body)

	shortURL := "EwHXdJfB" //generateShortURL()
	urlStore[shortURL] = url

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "http://localhost:8080/%s\n", shortURL)
}

func RedirectURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[1:]

	originalURL, err := urlStore[id]
	if !err {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Header().Set("Location", originalURL)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", ShortenURL)
	mux.HandleFunc("/{id}", RedirectURL)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
