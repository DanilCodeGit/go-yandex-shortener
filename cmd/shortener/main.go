package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/DanilCodeGit/go-yandex-shortener/internal/cfg"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/handlers"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/logger"
	"github.com/go-chi/chi/v5"
)

func main() {
	err := cfg.Env()
	if err != nil {
		log.Fatal(err)
	}

	flag.Parse()

	r := chi.NewRouter()

	loggerGet := logger.WithLogging(http.HandlerFunc(handlers.HandleGet))
	loggerPost := logger.WithLogging(http.HandlerFunc(handlers.HandlePost))
	loggerJSONHandler := logger.WithLogging(http.HandlerFunc(handlers.JSONHandler))
	r.Get("/{id}", loggerGet.ServeHTTP)
	r.Post("/", loggerPost.ServeHTTP)
	r.Post("/api/shorten", loggerJSONHandler.ServeHTTP)

	err = http.ListenAndServe(*cfg.FlagServerAddress, r)
	if err != nil {
		panic(err)
	}
}
