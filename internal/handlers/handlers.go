package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

func saveDataToFile(data map[string]string, filePath string) error {
	// Преобразуем данные в формат URLData
	var jsonData []URLData
	for shortURL, originalURL := range data {
		jsonData = append(jsonData, URLData{ShortURL: shortURL, OriginalURL: originalURL})
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(jsonData); err != nil {
		return err
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

	//// Сохраняем данные в файл после обновления
	//if err := saveDataToFile(st, *cfg.FlagFileStoragePath); err != nil {
	//	http.Error(w, "Failed to save data to file", http.StatusInternalServerError)
	//	return
	//}

	// Полный путь к файлу базы данных JSON
	dbFilePath, err := filepath.Abs(*cfg.FlagFileStoragePath)
	if err != nil {
		http.Error(w, "Failed to get absolute path for the database file", http.StatusInternalServerError)
		return
	}

	// Сохраняем данные в файл после обновления
	if err := saveDataToFile(st, dbFilePath); err != nil {
		http.Error(w, "Failed to save data to file", http.StatusInternalServerError)
		return
	}

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
	responseData := map[string]string{"result": shortURL}
	responseJSON, _ := json.Marshal(responseData)

	// Сохраняем данные в файл после обновления
	if err := saveDataToFile(st, *cfg.FlagFileStoragePath); err != nil {
		http.Error(w, "Failed to save data to file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v", string(responseJSON))

}
