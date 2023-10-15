// Добавить функцию сохранения json в файле
// Добавить тесты
// Исправить storage (Использовать структуру storage)

// Работающий вариант
package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/DanilCodeGit/go-yandex-shortener/internal/cfg"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/database/postgre"
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

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		_ = fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(jsonData); err != nil {
		return err
	}

	return nil
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

	// Преобразование данных в формат JSON
	jsonData := make(map[string]string)
	mu.Lock()
	for shortURL, originalURL := range st {
		jsonData[shortURL] = originalURL
	}
	mu.Unlock()

	////////////////////// DATABASE
	conn, err := postgre.DBConn()
	if err != nil {
		log.Println("Неудачное подключение")
	}
	err = postgre.CreateTable(conn)
	if err != nil {
		log.Println("База не создана")
	}

	err = postgre.SaveShortenedURL(conn, st[ShortURL], ShortURL)
	if err != nil {
		log.Println("Запись не произошла")
	}
	///////////////////////

	// Сохранение данных в файл после обновления
	err = saveDataToFile(jsonData, *cfg.FlagFileStoragePath)
	if err != nil {
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
	originalURL := st[shortURL]
	delete(st, "url")

	shortURL = "http://localhost:8080" + "/" + shortURL
	responseData := map[string]string{"result": shortURL}
	responseJSON, _ := json.Marshal(responseData)

	jsonData := make(map[string]string)
	mu.Lock()
	for shortURL, originalURL := range st {
		jsonData[shortURL] = originalURL
	}
	mu.Unlock()

	// Сохранение данных в файл после обновления
	err = saveDataToFile(jsonData, *cfg.FlagFileStoragePath)
	if err != nil {
		http.Error(w, "Failed to save data to file", http.StatusInternalServerError)
		return
	}
	////////////////////// DATABASE
	conn, err := postgre.DBConn()
	if err != nil {
		log.Println("Неудачное подключение")
	}
	err = postgre.CreateTable(conn)
	if err != nil {
		log.Println("База не создана")
	}

	err = postgre.SaveShortenedURL(conn, originalURL, shortURL)
	if err != nil {
		log.Println("Запись не произошла")
	}
	///////////////////////

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v", string(responseJSON))

}

type Multi struct {
	CorrelationID string `json:"correlation_Id"`
	OriginalURL   string `json:"original_url"`
}

func MultipleRequestHandler(w http.ResponseWriter, r *http.Request) {
	var m []Multi
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	type ShortenStruct struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}

	var shortenData []ShortenStruct

	for _, item := range m {
		// Создаем хеш SHA-256 от OriginalUrl
		hash := tools.HashURL(item.OriginalUrl)
		shortURL := hash
		shortenData = append(shortenData, ShortenStruct{
			CorrelationId: item.CorrelationId,
			ShortUrl:      shortURL,
		})
	}

	shortenJSON, err := json.Marshal(shortenData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(shortenJSON)
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

func HandlePing(w http.ResponseWriter, r *http.Request) {
	conn, err := postgre.DBConn()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatalf("Хэндлер не может подключиться к бд")
	}
	defer conn.Close()
	w.Header().Set("Location", "Success")
	w.WriteHeader(http.StatusOK)

}
