package main

import (
	"context"
	_ "database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/DanilCodeGit/go-yandex-shortener/cmd/shortener/gzip"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/auth"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/cfg"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/database/postgre"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/handlers"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/logger"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg.InitConfig()
	ctx := context.Background()
	conn, err := postgre.NewDataBase(context.Background(), *cfg.FlagDataBaseDSN)
	if err != nil {
		log.Fatal("Database connection failed")
	}
	err = conn.CreateTable(ctx)
	if err != nil {
		log.Println("Таблица уже создана")
	}

	r := chi.NewRouter()
	GetPing := logger.WithLogging(auth.AuthMiddleWare(gzipMiddleware(handlers.HandlePing(conn))))
	Get := logger.WithLogging(auth.AuthMiddleWare(gzipMiddleware(handlers.HandleGet())))
	Post := logger.WithLogging(auth.AuthMiddleWare(gzipMiddleware(handlers.HandlePost(conn))))
	JSONHandler := logger.WithLogging(auth.AuthMiddleWare(gzipMiddleware(handlers.JSONHandler(conn))))
	MultipleRequestHandler := logger.WithLogging(auth.AuthMiddleWare(gzipMiddleware(handlers.MultipleRequestHandler(conn))))
	GetUserURLs := logger.WithLogging(auth.AuthMiddleWare(handlers.GetUserURLs()))
	r.Get("/api/user/urls", GetUserURLs.ServeHTTP)
	r.Get("/ping", GetPing.ServeHTTP)
	r.Get("/{id}", Get.ServeHTTP)
	r.Post("/", Post.ServeHTTP)
	r.Post("/api/shorten", JSONHandler.ServeHTTP)
	r.Post("/api/shorten/batch", MultipleRequestHandler.ServeHTTP)

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
