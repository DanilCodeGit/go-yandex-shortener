package cfg

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
)

// Переменные окружения
type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

func Env() {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
}

// Флаги
var (
	FlagServerAddress = flag.String("a", "localhost:8080", "Адрес запуска HTTP-сервера")
	FlagBaseURL       = flag.String("b", "http://localhost:8080", "Базовый адрес результирующего сокращённого URL")
)
