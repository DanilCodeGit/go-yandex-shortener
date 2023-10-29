package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/DanilCodeGit/go-yandex-shortener/internal/auth"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/cfg"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/database/postgre"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/storage"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/tools"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgerrcode"
)

var st = *storage.NewStorage()

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

func HandlePost(db *postgre.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jwt")
		if err != nil || cookie.Value == "" {
			log.Fatal("Траблы с куки", err)
			return
		}
		userID := auth.GetUserID(cookie.Value)
		st.UserID = userID
		ctx := r.Context()
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

		st.SetURL(shortURL, url)
		// Преобразование данных в формат JSON
		jsonData := make(map[string]string)

		for shortURL, originalURL := range st.URLsStore {
			jsonData[shortURL] = originalURL
		}
		////////////////////// DATABASE

		code, _ := db.SaveShortenedURL(ctx, url, shortURL)
		if code == pgerrcode.UniqueViolation {
			log.Println("Запись не произошла")
			w.WriteHeader(http.StatusConflict)
			fprintf, err := fmt.Fprintf(w, "%s/%s", *cfg.FlagBaseURL, shortURL)
			if err != nil {
				return
			}
			fmt.Print(fprintf)
			return
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
		fprintf, err := fmt.Fprintf(w, "%s/%s", *cfg.FlagBaseURL, shortURL)
		if err != nil {
			return
		}
		fmt.Print(fprintf)
	}
}

func JSONHandler(db *postgre.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		cookie, err := req.Cookie("jwt")
		if err != nil || cookie.Value == "" {
			http.Error(w, "Необходима аутентификация", http.StatusUnauthorized)
			return
		}
		userID := auth.GetUserID(cookie.Value)
		st.UserID = userID
		ctx := req.Context()
		var buf bytes.Buffer
		// читаем тело запроса
		_, err = buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Десереализуем json
		if err = json.Unmarshal(buf.Bytes(), &st.URLsStore); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		url, found := st.GetURL("url")
		if !found {
			http.Error(w, "Missing 'url' field in JSON", http.StatusBadRequest)
			return
		}
		shortURL := tools.HashURL(url)

		st.SetURL(shortURL, url)
		originalURL, _ := st.GetURL(shortURL)
		fmt.Println("original: ", originalURL)

		st.DeleteURL("url")

		shortURL = "http://localhost:8080" + "/" + shortURL
		responseData := map[string]string{"result": shortURL}
		responseJSON, _ := json.Marshal(responseData)

		newData := make(map[string]string)

		for shortURL, originalURL := range st.URLsStore {
			newData[shortURL] = originalURL
		}

		// Сохранение данных в файл после обновления
		err = saveDataToFile(newData, *cfg.FlagFileStoragePath)
		if err != nil {
			http.Error(w, "Failed to save data to file", http.StatusInternalServerError)
			return
		}
		////////////////////// DATABASE

		code, _ := db.SaveShortenedURL(ctx, url, shortURL)
		if code == pgerrcode.UniqueViolation {
			log.Println("Запись не произошла")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			fprintf, err := fmt.Fprintf(w, "%v", string(responseJSON))
			if err != nil {
				return
			}
			fmt.Println(fprintf)
			return
		}

		///////////////////////

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fprintf, err := fmt.Fprintf(w, "%v", string(responseJSON))
		if err != nil {
			return
		}
		fmt.Println(fprintf)
	}
}

type Multi struct {
	CorrelationID string `json:"correlation_Id"`
	OriginalURL   string `json:"original_url"`
}

func MultipleRequestHandler(db *postgre.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jwt")
		if err != nil || cookie.Value == "" {
			http.Error(w, "Необходима аутентификация", http.StatusUnauthorized)
			return
		}
		userID := auth.GetUserID(cookie.Value)
		st.UserID = userID
		ctx := r.Context()
		var m []Multi
		var buf bytes.Buffer

		_, err = buf.ReadFrom(r.Body)
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
			hash := tools.HashURL(item.OriginalURL)
			tmp := hash
			shortURL := "http://localhost:8080" + "/" + hash
			st.SetBatch(tmp, item.OriginalURL)

			shortenData = append(shortenData, ShortenStruct{
				CorrelationID: item.CorrelationID,
				ShortURL:      shortURL,
			})

		}

		shortenJSON, err := json.Marshal(shortenData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		newData := make(map[string]string)

		for shortURL, originalURL := range st.URLsStore {
			newData[shortURL] = originalURL
		}

		// Сохранение данных в файл после обновления
		err = saveDataToFile(newData, *cfg.FlagFileStoragePath)
		if err != nil {
			http.Error(w, "Failed to save data to file", http.StatusInternalServerError)
			return
		}

		////////////////////// DATABASE

		for shortURL, originalURL := range newData {
			code, _ := db.SaveBatch(ctx, originalURL, shortURL)
			if code == pgerrcode.UniqueViolation {
				log.Println("Запись не произошла")
				w.WriteHeader(http.StatusConflict)
				fprintf, err := fmt.Fprintf(w, "%s/%s", *cfg.FlagBaseURL, shortURL)
				if err != nil {
					return
				}
				fmt.Print(fprintf)
				return
			}

		}

		///////////////////////

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(shortenJSON)
	}
}

func HandleGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		originalURL, ok := st.GetURL(id)
		if !ok {
			log.Fatal(originalURL)
		}

		fmt.Println("orig: ", originalURL)
		fmt.Println("st: ", st.URLsStore)
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func HandlePing(db *postgre.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer r.Body.Close()
		_, err := postgre.NewDataBase(context.Background(), *cfg.FlagDataBaseDSN)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatalf("Хэндлер не может подключиться к бд")
		}
		db.Close(ctx)
		w.Header().Set("Location", "Success")
		w.WriteHeader(http.StatusOK)
	}
}

var userURLs = map[int][]*storage.Storage{}

func GetUserURLs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получить куку JWT из запроса
		cookie, err := r.Cookie("jwt")
		if err != nil || cookie.Value == "" {
			http.Error(w, "Необходима аутентификация", http.StatusUnauthorized)
			return
		}

		// Извлечь UserID из куки
		userID := auth.GetUserID(cookie.Value)
		if userID == -1 {
			http.Error(w, "Недействительный JWT-токен", http.StatusUnauthorized)
			return
		}
		userStorage := &storage.Storage{
			URLsStore: st.URLsStore,
			UserID:    userID, // Присвойте здесь идентификатор пользователя.
		}
		userURLs[userID] = append(userURLs[userID], userStorage)
		// Поиск сокращенных URL для данного пользователя
		urls, exists := userURLs[userID]

		var formatURLs []URLData
		for _, userStorage := range urls {
			for shortURL, originalURL := range userStorage.URLsStore {
				formatURLs = append(formatURLs, URLData{
					ShortURL:    shortURL,
					OriginalURL: originalURL,
				})
			}
		}
		log.Println("formatURLs: ", formatURLs)

		if !exists || len(urls) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		log.Println(userURLs)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(formatURLs)

	}
}
