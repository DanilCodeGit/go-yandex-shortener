package handlers

import (
	"bufio"
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

type ProducerData struct {
	file   *os.File // файл для записи
	writer *bufio.Writer
}

func NewProducerData() (*ProducerData, error) {
	// Создаем временный файл в каталоге /tmp/
	tmpFile, err := os.CreateTemp("tmp", "*.json")
	if err != nil {
		return nil, err
	}

	return &ProducerData{
		file:   tmpFile,
		writer: bufio.NewWriter(tmpFile),
	}, nil
}

func (p *ProducerData) Close() error {
	// закрываем файл
	return p.file.Close()
}

func (p *ProducerData) WriteEvent(urls *URLData) error {
	data, err := json.Marshal(&urls)
	if err != nil {
		return err
	}

	// записываем событие в буфер
	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	// добавляем перенос строки
	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	// записываем буфер в файл
	return p.writer.Flush()
}

type Consumer struct {
	file *os.File
	// добавляем reader в Consumer
	scanner *bufio.Scanner
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file: file,
		// создаём новый Reader
		scanner: bufio.NewScanner(file),
	}, nil
}

func (c *Consumer) ReadEvent() (*URLData, error) {
	// одиночное сканирование до следующей строки
	if !c.scanner.Scan() {
		return nil, c.scanner.Err()
	}
	// читаем данные из scanner
	data := c.scanner.Bytes()

	urls := URLData{}
	err := json.Unmarshal(data, &urls)
	if err != nil {
		return nil, err
	}

	return &urls, nil
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

	// Сохранить в файл, если указан путь для хранения данных
	if *cfg.FlagFileStoragePath != "" {
		p, err := NewProducerData()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer p.Close()
		// Записать в файл
		if err := p.WriteEvent(&URLData{ShortURL: ShortURL, OriginalURL: url}); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
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

	mu.Lock()
	st["result"] = shortURL
	mu.Unlock()

	if *cfg.FlagFileStoragePath != "" {
		p, err := NewProducerData()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer p.Close()
		// Записать в файл
		if err := p.WriteEvent(&URLData{ShortURL: shortURL, OriginalURL: url}); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	responseData := map[string]string{"result": shortURL}
	responseJSON, _ := json.Marshal(responseData)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%v", string(responseJSON))

}
