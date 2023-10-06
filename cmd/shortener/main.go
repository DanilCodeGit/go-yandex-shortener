package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/DanilCodeGit/go-yandex-shortener/cmd/shortener/gzip"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/cfg"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/handlers"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/logger"
	"github.com/go-chi/chi/v5"
)

func gzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		ow := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := gzip.NewCompressWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			ow = cw
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer cw.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := gzip.NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer cr.Close()
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	}
}

func main() {
	// Имя переменной окружения, которую вы хотите установить
	envVarName := "FILE_STORAGE_PATH"

	// Значение, которое вы хотите установить для переменной окружения
	envVarValue := "/tmp/short-url-db.json"

	// Установить переменную окружения
	err := os.Setenv(envVarName, envVarValue)

	// Проверить наличие ошибки при установке переменной окружения
	if err != nil {
		fmt.Printf("Ошибка при установке переменной окружения %s: %v\n", envVarName, err)
	} else {
		fmt.Printf("Переменная окружения %s успешно установлена\n", envVarName)
	}

	handlers.Init()
	err = cfg.Env()
	if err != nil {
		log.Fatal(err)
	}

	flag.Parse()

	r := chi.NewRouter()

	loggerGet := logger.WithLogging(gzipMiddleware(handlers.HandleGet))
	loggerPost := logger.WithLogging(gzipMiddleware(handlers.HandlePost))
	loggerJSONHandler := logger.WithLogging(gzipMiddleware(handlers.JSONHandler))

	r.Get("/{id}", loggerGet.ServeHTTP)
	r.Post("/", loggerPost.ServeHTTP)
	r.Post("/api/shorten", loggerJSONHandler.ServeHTTP)

	err = http.ListenAndServe(*cfg.FlagServerAddress, r)
	if err != nil {
		panic(err)
	}
}
