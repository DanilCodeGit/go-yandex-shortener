package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/DanilCodeGit/go-yandex-shortener/internal/cfg"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/storage"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/tools"
)

var st = storage.URLStore
var mu sync.Mutex

type URLData struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func saveURLsToDisk(filePath string, urls map[string]string) error {
	var urlData []URLData

	for shortURL, originalURL := range urls {
		urlData = append(urlData, URLData{
			ShortURL:    shortURL,
			OriginalURL: originalURL,
		})
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(urlData)
	if err != nil {
		return err
	}

	return nil
}

func loadURLsFromDisk(filePath string, urls map[string]string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var urlData []URLData
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&urlData)
	if err != nil {
		return err
	}

	for _, data := range urlData {
		urls[data.ShortURL] = data.OriginalURL
	}

	return nil
}

func HandleGet(w http.ResponseWriter, r *http.Request) {
	// Разбить путь запроса на части
	parts := strings.Split(r.URL.Path, "/")

	// Извлечь значение {id}
	if len(parts) < 2 || parts[1] == "" {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}
	id := parts[1]
	mu.Lock()
	originalURL := storage.URLStore[id]
	mu.Unlock()
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func HandlePost(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if len(body) == 0 {
		http.Error(w, "Тело запроса пустое", http.StatusBadRequest)
		return
	}

	url := string(body)

	ShortURL := tools.HashURL(url)
	mu.Lock()
	st[ShortURL] = url
	mu.Unlock()
	//
	if *cfg.FlagFileStoragePath != "" {
		err := saveURLsToDisk(*cfg.FlagFileStoragePath, st)
		if err != nil {
			http.Error(w, "Failed to save data to disk", http.StatusInternalServerError)
			return
		}
	}
	//
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fprintf, err := fmt.Fprintf(w, "%s/%s", *cfg.FlagBaseURL, ShortURL)
	if err != nil {
		return
	}
	fmt.Print(fprintf)
}

func JSONHandler(w http.ResponseWriter, req *http.Request) { //POST

	var buf bytes.Buffer
	// читаем тело запроса
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Десереализуем json
	if err = json.Unmarshal(buf.Bytes(), &st); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	url, found := st["url"]
	if !found {
		http.Error(w, "Missing 'url' field in JSON", http.StatusBadRequest)
		return
	}
	shortURL := tools.HashURL(url)
	st[shortURL] = url
	shortURL = "http://localhost:8080" + "/" + shortURL

	mu.Lock()
	st["result"] = shortURL
	mu.Unlock()

	responseData := map[string]string{"result": shortURL}
	responseJSON, _ := json.Marshal(responseData)
	// Загрузка данных URL с диска
	if *cfg.FlagFileStoragePath != "" {
		err := loadURLsFromDisk(*cfg.FlagFileStoragePath, st)
		if err != nil {
			http.Error(w, "Failed to load data from disk", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s", string(responseJSON))
}
