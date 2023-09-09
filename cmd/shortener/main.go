package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var urlStore = make(map[string]string)

func handleROOT(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		handleGet(w, r)
	}
	if r.Method == http.MethodPost {
		handlePost(w, r)
	}
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Разбить путь запроса на части
		parts := strings.Split(r.URL.Path, "/")

		// Извлечь значение {id}
		if len(parts) < 2 || parts[1] == "" {
			http.Error(w, "Некорректный запрос", http.StatusBadRequest)
			return
		}
		id := parts[1]
		// Ваша логика для получения оригинального URL на основе id.
		originalURL := urlStore[id]
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
}
func handlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Read the URL from the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Convert the request body to a string
		url := string(body)

		shortURL := hashURL(url) //generateShortURL()
		urlStore[shortURL] = url
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)

		fmt.Fprintf(w, "http://localhost:8080/%s", shortURL)
	} else {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
}

func hashURL(urlToHash string) string {
	// Создайте новый хеш SHA-256
	hasher := sha256.New()

	// Преобразуйте URL в байтовый срез и передайте его хешеру
	hasher.Write([]byte(urlToHash))

	// Получите байтовое представление хеша
	hashBytes := hasher.Sum(nil)

	// Преобразуйте байты хеша в строку в шестнадцатеричном формате
	hashedURL := hex.EncodeToString(hashBytes)

	return hashedURL
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, handleROOT)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
