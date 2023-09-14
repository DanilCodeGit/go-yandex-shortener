package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// var (
//
//	listenAddr = flag.String("a", "localhost:8080", "Адрес запуска HTTP-сервера")
//	baseURL    = flag.String("b", "http://localhost:8080", "Базовый адрес результирующего сокращённого URL")
//
// )
var (
	listenAddr = flag.String("a", "", "Адрес запуска HTTP-сервера")
	baseURL    = flag.String("b", "", "Базовый адрес результирующего сокращённого URL")
)
var urlStore = make(map[string]string)

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

func handleGet(w http.ResponseWriter, r *http.Request) {
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
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	// Read the URL from the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if len(body) == 0 {
		http.Error(w, "Тело запроса пустое", http.StatusBadRequest)
		return
	}
	// Convert the request body to a string
	url := string(body)

	ShortURL := hashURL(url) //generateShortURL()
	urlStore[ShortURL] = url
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s/%s", *baseURL, ShortURL)
	//fmt.Fprintf(w, "http://localhost:8080/%s", ShortURL)
}

func main() {
	r := chi.NewRouter()
	r.Get("/{id}", handleGet)
	r.Post("/", handlePost)
	err := http.ListenAndServe(*listenAddr, r)
	if err != nil {
		panic(err)
	}
}
