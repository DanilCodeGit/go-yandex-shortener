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
	loggerGet := logger.WithLogging(http.HandlerFunc(handlers.HandleGet))
	loggerPost := logger.WithLogging(http.HandlerFunc(handlers.HandlePost))
	flag.Parse()
	r := chi.NewRouter()
	r.Get("/{id}", loggerGet.ServeHTTP)
	r.Post("/", loggerPost.ServeHTTP)
	err = http.ListenAndServe(*cfg.FlagServerAddress, r)
	if err != nil {
		panic(err)
	}
}
