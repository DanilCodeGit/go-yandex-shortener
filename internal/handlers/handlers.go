package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/DanilCodeGit/go-yandex-shortener/internal/cfg"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/storage"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/tools"
)

var st = storage.URLStore
var mu sync.Mutex

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
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fprintf, err := fmt.Fprintf(w, "%s/%s", *cfg.FlagBaseURL, ShortURL)
	if err != nil {
		return
	}
	fmt.Print(fprintf)
}

func JSONHandler(w http.ResponseWriter, req *http.Request) {
	// Принять в тело json
	// Конвертировать в строку
	// Сократить
	// Отдать json

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
	shortURL = "http://localhost:8080" + "/" + shortURL
	st[shortURL] = url
	delete(st, "url")
	mu.Lock()
	st["result"] = shortURL
	mu.Unlock()

	responseData := map[string]string{"result": shortURL}
	responseJSON, _ := json.Marshal(responseData)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s", string(responseJSON))

}
