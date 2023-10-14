package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/DanilCodeGit/go-yandex-shortener/cmd/shortener/gzip"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/cfg"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/database/postgre"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/handlers"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/logger"
	"github.com/go-chi/chi/v5"
)

func main() {
	// Вызываем функцию подключения к базе данных
	conn, err := postgre.DBConn()
	if err != nil {
		log.Fatal(err)
	}
	//defer conn.Close(context.Background())

	// Вызываем функцию создания таблицы
	err = postgre.CreateTable(conn)
	if err != nil {
		log.Fatal(err)
	}

	cfg.InitConfig()

	r := chi.NewRouter()
	loggerGetPing := handlers.HandlePing
	loggerGet := logger.WithLogging(gzipMiddleware(handlers.HandleGet))
	loggerPost := logger.WithLogging(gzipMiddleware(handlers.HandlePost))
	loggerJSONHandler := logger.WithLogging(gzipMiddleware(handlers.JSONHandler))
	r.Get("/ping", loggerGetPing)
	r.Get("/{id}", loggerGet.ServeHTTP)
	r.Post("/", loggerPost.ServeHTTP)
	r.Post("/api/shorten", loggerJSONHandler.ServeHTTP)

	err = http.ListenAndServe(*cfg.FlagServerAddress, r)
	if err != nil {
		panic(err)
	}
}

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
