package cfg

import (
	"flag"
	"os"

	"github.com/caarlos0/env"
)

// Config Переменные окружения
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DataBaseDSN     string `env:"DATABASE_DSN"`
}

// Флаги
var (
	FlagDataBaseDSN     = flag.String("d", "host=localhost port=5434 database=postgres user=postgres password=dixedDIX-111 sslmode=disable", "Строка подключения к БД")
	FlagServerAddress   = flag.String("a", "localhost:8080", "Адрес запуска HTTP-сервера")
	FlagFileStoragePath = flag.String("f", "/tmp/short-url-db.json", "Полное имя файла до JSON")
	FlagBaseURL         = flag.String("b", "http://localhost:8080", "Базовый адрес результирующего сокращённого URL")
)

func InitConfig() (*Config, error) {
	flag.Parse() // Парсинг флагов

	var cfg Config
	err := env.Parse(&cfg) // Парсинг переменных окружения
	if err != nil {
		return nil, err
	}

	// Если переменные окружения существуют, установите флаги из переменных окружения
	if serverAddressEnv, exists := os.LookupEnv("SERVER_ADDRESS"); exists {
		*FlagServerAddress = serverAddressEnv
	}
	if fileStoragePathEnv, exists := os.LookupEnv("FILE_STORAGE_PATH"); exists {
		*FlagFileStoragePath = fileStoragePathEnv
	}
	if baseURLEnv, exists := os.LookupEnv("BASE_URL"); exists {
		*FlagBaseURL = baseURLEnv
	}
	if dataBaseDSNEnv, exists := os.LookupEnv("DATABASE_DSN"); exists {
		*FlagDataBaseDSN = dataBaseDSNEnv
	}

	return &cfg, nil
}
