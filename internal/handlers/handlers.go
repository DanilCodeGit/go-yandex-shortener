// Добавить тесты

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
	"time"

	"github.com/DanilCodeGit/go-yandex-shortener/internal/cfg"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/database/postgre"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/storage"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/tools"
	"github.com/go-chi/chi/v5"
)

type URLData struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

var st = storage.NewStorage()

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
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
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
	if *cfg.FlagDataBaseDSN != "" {
		conn, err := postgre.DBConn(ctx)
		if err != nil {
			log.Println("Неудачное подключение")
		}
		err = postgre.CreateTable(conn)
		if err != nil {
			log.Println("База не создана")
		}

		err = postgre.CheckDuplicate(ctx, conn, url)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			fprintf, err := fmt.Fprintf(w, "%s/%s", *cfg.FlagBaseURL, shortURL)
			if err != nil {
				return
			}
			fmt.Print(fprintf)
			return
		}

		err = postgre.SaveShortenedURL(conn, url, shortURL)
		if err != nil {
			log.Println("Запись не произошла")
		}
	}

	///////////////////////

	// Сохранение данных в файл после обновления
	err = saveDataToFile(jsonData, *cfg.FlagFileStoragePath)
	if err != nil {
		http.Error(w, "Failed to save data to file", http.StatusInternalServerError)
		return
	}
	fmt.Println(st.URLsStore)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fprintf, err := fmt.Fprintf(w, "%s/%s", *cfg.FlagBaseURL, shortURL)
	if err != nil {
		return
	}
	fmt.Print(fprintf)
}

func JSONHandler(w http.ResponseWriter, req *http.Request) { //POST
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
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
	st.SetURL(url, shortURL)

	st.DeleteURL("url")
	shortURLDB := shortURL
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
	fmt.Println("url:", url)
	fmt.Println("shortUrl:", shortURLDB)
	////////////////////// DATABASE
	if *cfg.FlagDataBaseDSN != "" {
		conn, err := postgre.DBConn(ctx)
		if err != nil {
			log.Println("Неудачное подключение")
		}
		err = postgre.CreateTable(conn)
		if err != nil {
			log.Println("База не создана")
		}

		err = postgre.CheckDuplicate(ctx, conn, url)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			fmt.Fprint(w, string(responseJSON))
			return
		}

		err = postgre.SaveShortenedURL(conn, url, shortURLDB)
		if err != nil {
			log.Println("Запись не произошла")
		}
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
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
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

		hash := tools.HashURL(item.OriginalURL)
		tmp := hash
		shortURL := "http://localhost:8080" + "/" + hash
		st.SetURL(tmp, item.OriginalURL)
		//st[tmp] = item.OriginalURL
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
	if *cfg.FlagDataBaseDSN != "" {
		conn, err := postgre.DBConn(ctx)
		if err != nil {
			log.Println("Неудачное подключение")
		}
		err = postgre.CreateTable(conn)
		if err != nil {
			log.Println("База не создана")
		}

		for shortURL, originalURL := range newData {
			err = postgre.SaveShortenedURL(conn, originalURL, shortURL)
			if err != nil {
				log.Println("Запись не произошла")
			}
		}
	}

	///////////////////////

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(shortenJSON)
}

func HandleGet(w http.ResponseWriter, r *http.Request) {
	//// Разбить путь запроса на части
	//parts := strings.Split(r.URL.Path, "/")
	//
	//// Извлечь значение {id}
	//if len(parts) < 2 || parts[1] == "" {
	//	http.Error(w, "Некорректный запрос", http.StatusBadRequest)
	//	return
	//}
	//shortURL := parts[1]
	//fmt.Println("GET st: ", st.URLsStore)
	//fmt.Println("shortURL: ", shortURL)
	//
	//// Найти соответствующий originalURL
	//originalURL, ok := st.URLsStore[shortURL]
	//if !ok {
	//	log.Println("OriginalURL not found")
	//	w.WriteHeader(http.StatusNotFound)
	//	//return
	//}
	//
	//// Выполнить редирект на originalURL
	//w.Header().Set("Location", originalURL)
	//w.WriteHeader(http.StatusTemporaryRedirect)
	//fmt.Println("originalURLGET:", originalURL)

	shortURL := chi.URLParam(r, "id")
	// Удалить записи из базы данных перед обработкой запроса
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	conn, err := postgre.DBConn(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatalf("Хэндлер не может подключиться к бд")
	}
	defer conn.Close()

	// Удаление записей из базы данных
	err = postgre.DeleteAllRecords(conn)
	if err != nil {
		log.Printf("Ошибка при удалении записей из базы данных: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	origURL, ok := st.GetURL(shortURL)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", origURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
	fmt.Println("id:", shortURL)
}

func HandlePing(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	conn, err := postgre.DBConn(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatalf("Хэндлер не может подключиться к бд")
	}
	defer conn.Close()
	w.Header().Set("Location", "Success")
	w.WriteHeader(http.StatusOK)

}
