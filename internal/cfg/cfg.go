package cfg

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env"
)

// Config Переменные окружения
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func Env() error {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		fmt.Println("Невозможно спарсить cfg")
	}
	return err
}

// Флаги
var (
	FlagServerAddress   = flag.String("a", "localhost:8080", "Адрес запуска HTTP-сервера")
	FlagFileStoragePath = flag.String("f", `short-url-db.json`, "Полное имя файла до JSON")
	FlagBaseURL         = flag.String("b", "http://localhost:8080", "Базовый адрес результирующего сокращённого URL")
)
