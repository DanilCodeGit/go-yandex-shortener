package main

import (
	"fmt"
	"io"
	"net/http"
)

var urlStore = make(map[string]string)

func ShortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "It's not a post-request", http.StatusMethodNotAllowed)
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
	fmt.Fprintf(w, "http://localhost:8080/%s", shortURL)
}

func RedirectURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	// Получаем значение параметра {id} из URL
	id := r.URL.Path[1:] // Убираем первый символ ("/")

	// Здесь вы можете добавить логику для сопоставления id с оригинальным URL.
	// Например, вы можете использовать карту (map) для хранения соответствий.

	// Если id не найден, вернем код 400
	if id == "" {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	// Получите оригинальный URL на основе id (замените на вашу логику).

	// В случае успешной обработки запроса, вернем код 307 и оригинальный URL
	originalURL := "https://practicum.yandex.ru/" // Замените на ваш оригинальный URL
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, ShortenURL)
	mux.HandleFunc(`/EwHXdJfB`, RedirectURL)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
