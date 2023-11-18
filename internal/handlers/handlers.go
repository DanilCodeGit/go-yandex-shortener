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
	"sync"

	"github.com/DanilCodeGit/go-yandex-shortener/internal/auth"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/cfg"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/database/postgre"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/storage"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/tools"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgerrcode"
)

var st = *storage.NewStorage()

type DataBaseHandle struct {
	DB *postgre.DB
}
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

func (db *DataBaseHandle) HandlePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

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

		code, _ := db.DB.SaveShortenedURL(ctx, url, shortURL)
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

func (db *DataBaseHandle) JSONHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		ctx := req.Context()
		var buf bytes.Buffer
		// читаем тело запроса
		_, err := buf.ReadFrom(req.Body)
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

		code, _ := db.DB.SaveShortenedURL(ctx, url, shortURL)
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

func (db *DataBaseHandle) MultipleRequestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()
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
		code, _ := db.DB.SaveBatch(ctx, newData)
		if code == pgerrcode.UniqueViolation {
			log.Println("Запись не произошла")
			w.WriteHeader(http.StatusConflict)
			w.Write(shortenJSON)
			return
		}

		///////////////////////

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(shortenJSON)
	}
}

func (db *DataBaseHandle) HandleGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		originalURL, ok := st.GetURL(id)
		if !ok {
			log.Fatal(originalURL)
		}
		flag, err := db.DB.GetFlagShortURL(id)
		if err != nil {
			log.Println("Ошибка получения shortURL из запроса к бд")
		}
		if flag {
			w.WriteHeader(http.StatusGone)
			return
		}
		fmt.Println("orig: ", originalURL)
		fmt.Println("st: ", st.URLsStore)
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func (db *DataBaseHandle) HandlePing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := r.Context()
		_, err := postgre.NewDataBase(context.Background(), *cfg.FlagDataBaseDSN)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatalf("Хэндлер не может подключиться к бд")
		}
		db.DB.Close(ctx)
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
			http.Error(w, "Необходима аутентификация", http.StatusNoContent)
			return
		}

		// Извлечь UserID из куки
		userID := auth.GetUserID(cookie.Value)
		if userID == -1 {
			http.Error(w, "Недействительный JWT-токен", http.StatusUnauthorized)
			return
		}
		userStorage := storage.Storage{
			URLsStore: st.URLsStore,
			UserID:    userID,
		}
		userURLs[userID] = append(userURLs[userID], &userStorage)
		// Поиск сокращенных URL для данного пользователя
		urls, exists := userURLs[userID]
		if !exists || len(urls) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var formatURLs []URLData
		for _, userStorage := range urls {
			for shortURL, originalURL := range userStorage.URLsStore {
				formatURLs = append(formatURLs, URLData{
					ShortURL:    "http://localhost:8080/" + shortURL,
					OriginalURL: originalURL,
				})
			}
			//break
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(formatURLs)

	}
}

func (db *DataBaseHandle) DeleteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получить куку JWT из запроса
		cookie, err := r.Cookie("jwt")
		if err != nil || cookie.Value == "" {
			http.Error(w, "Необходима аутентификация", http.StatusNoContent)
			return
		}

		// Извлечь UserID из куки
		userID := auth.GetUserID(cookie.Value)
		if userID == -1 {
			http.Error(w, "Недействительный JWT-токен", http.StatusUnauthorized)
			return
		}
		// Получить куку JWT из запроса
		cookie, err = r.Cookie("jwt")
		if err != nil || cookie.Value == "" {
			http.Error(w, "Необходима аутентификация", http.StatusNoContent)
			return
		}

		// Извлечь UserID из куки
		userID = auth.GetUserID(cookie.Value)
		if userID == -1 {
			http.Error(w, "Недействительный JWT-токен", http.StatusUnauthorized)
			return
		}

		// Читаем тело запроса
		var urlsToDelete []string
		err = json.NewDecoder(r.Body).Decode(&urlsToDelete)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// канал для сигнала к выходу из горутины
		doneCh := make(chan struct{})
		// при завершении программы закрываем канал doneCh, чтобы все горутины тоже завершились
		defer close(doneCh)
		inputCh := generator(doneCh, urlsToDelete)
		log.Println("InputCh: ", inputCh)

		channels := fanOut(doneCh, inputCh, db.DB)
		result := fanIn(doneCh, channels...)

		for res := range result {
			log.Println(res)
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func generator(doneCh chan struct{}, input []string) chan string {
	inputCh := make(chan string)

	go func() {
		defer close(inputCh)

		for _, data := range input {
			select {
			case <-doneCh:
				return
			case inputCh <- data:
			}
		}
	}()

	return inputCh
}

// fanOut принимает канал данных, порождает 10 горутин
func fanOut(doneCh chan struct{}, inputCh chan string, db *postgre.DB) []chan error {
	// количество горутин add
	numWorkers := 20
	// каналы, в которые отправляются результаты
	channels := make([]chan error, numWorkers)

	for i := 0; i < numWorkers; i++ {
		// получаем канал из горутины add
		addResultCh := deleteURLs(doneCh, inputCh, db)
		// отправляем его в слайс каналов
		channels[i] = addResultCh
	}

	// возвращаем слайс каналов
	return channels
}

// fanIn объединяет несколько каналов resultChs в один.
func fanIn(doneCh chan struct{}, resultChs ...chan error) chan error {
	// конечный выходной канал в который отправляем данные из всех каналов из слайса, назовём его результирующим
	finalCh := make(chan error)

	// понадобится для ожидания всех горутин
	var wg sync.WaitGroup

	// перебираем все входящие каналы
	for _, ch := range resultChs {
		// в горутину передавать переменную цикла нельзя, поэтому делаем так
		chClosure := ch

		// инкрементируем счётчик горутин, которые нужно подождать
		wg.Add(1)

		go func() {
			// откладываем сообщение о том, что горутина завершилась
			defer wg.Done()

			// получаем данные из канала
			for data := range chClosure {
				select {
				// выходим из горутины, если канал закрылся
				case <-doneCh:
					return
				// если не закрылся, отправляем данные в конечный выходной канал
				case finalCh <- data:
				}
			}
		}()
	}

	go func() {
		// ждём завершения всех горутин
		wg.Wait()
		// когда все горутины завершились, закрываем результирующий канал
		close(finalCh)
	}()

	// возвращаем результирующий канал
	return finalCh
}
func deleteURLs(doneCh chan struct{}, inputCh chan string, db *postgre.DB) chan error {
	addRes := make(chan error)

	go func() {
		defer close(addRes)

		for data := range inputCh {

			err := db.MarkURLAsDeleted(data)
			if err != nil {
				// Обработка ошибок при удалении
				log.Printf("Failed to mark URL %s as deleted: %v", data, err)
			}

			select {
			case <-doneCh:
				return
			case addRes <- err:
			}
		}
	}()
	return addRes
}
