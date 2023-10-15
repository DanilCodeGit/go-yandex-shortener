// Добавить функцию сохранения json в файле
// Добавить тесты
// Исправить storage (Использовать структуру storage)

package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	//"strings"
	"sync"

	"github.com/DanilCodeGit/go-yandex-shortener/internal/cfg"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/database/postgre"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/storage"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/tools"
	"github.com/go-chi/chi/v5"
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
		return err
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

	shortURL := tools.HashURL(url)

	st[shortURL] = url

	// Преобразование данных в формат JSON
	jsonData := make(map[string]string)
	mu.Lock()
	for shortURL, originalURL := range st {
		jsonData[shortURL] = originalURL
	}
	mu.Unlock()

	conn, err := postgre.DBConn()
	if err != nil {
		log.Println("Неудачное подключение")
	}

	err = postgre.CreateTable(conn)
	if err != nil {
		log.Println("База не создана")
	}
	// Сохраняем запрос в бд
	err = postgre.SaveShortenedURL(conn, st[shortURL], shortURL)
	if err != nil {
		log.Println("Запись не произошла")
	}

	// Сохранение данных в файл после обновления
	err = saveDataToFile(jsonData, *cfg.FlagFileStoragePath)
	if err != nil {
		http.Error(w, "Failed to save data to file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fprintf, err := fmt.Fprintf(w, "%s/%s", *cfg.FlagBaseURL, shortURL)
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
	var url string
	for k, fullURL := range st {
		delete(st, k)
		url = fullURL
	}

	shortURL := tools.HashURL(url)
	st[shortURL] = url

	newData := make(map[string]string)
	mu.Lock()
	var full string
	for shortURL, originalURL := range st {
		newData[shortURL] = originalURL
		full = originalURL
	}

	shortURL = "http://localhost:8080" + "/" + shortURL
	responseData := map[string]string{"result": shortURL}
	responseJSON, _ := json.Marshal(responseData)

	mu.Unlock()
	// Сохранение данных в файл после обновления
	err = saveDataToFile(newData, *cfg.FlagFileStoragePath)
	if err != nil {
		http.Error(w, "Failed to save data to file", http.StatusInternalServerError)
		return
	}
	////
	conn, err := postgre.DBConn()
	if err != nil {
		log.Println("Неудачное подключение")
	}

	err = postgre.CreateTable(conn)
	if err != nil {
		log.Println("База не создана")
	}
	// Сохраняем запрос в бд
	err = postgre.SaveShortenedURL(conn, full, shortURL)
	if err != nil {
		log.Println("Запись не произошла")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v", string(responseJSON))

}

func HandleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	//// Разбить путь запроса на части
	//parts := strings.Split(r.URL.Path, "/")
	//
	//// Извлечь значение {id}
	//if len(parts) < 2 || parts[1] == "" {
	//	http.Error(w, "Некорректный запрос", http.StatusBadRequest)
	//	return
	//}
	//id := parts[1]
	mu.Lock()
	originalURL := storage.URLStore[id]
	mu.Unlock()

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
	fmt.Sprintf("Оригинальный url: %s\n", originalURL)

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
