package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/DanilCodeGit/go-yandex-shortener/internal/cfg"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/handlers"
	"github.com/go-chi/chi/v5"
)

func main() {
	err := cfg.Env()
	if err != nil {
		log.Fatal(err)
	}
	flag.Parse()
	r := chi.NewRouter()
	r.Get("/{id}", handlers.HandleGet)
	r.Post("/", handlers.HandlePost)
	err = http.ListenAndServe(*cfg.FlagServerAddress, r)
	if err != nil {
		panic(err)
	}
}
