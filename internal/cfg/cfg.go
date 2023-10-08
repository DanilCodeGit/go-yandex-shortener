// package cfg
//
// import (
//
//	"flag"
//	"fmt"
//	"os"
//
//	"github.com/caarlos0/env"
//
// )
//
// // Config Переменные окружения
//
//	type Config struct {
//		ServerAddress   string `env:"SERVER_ADDRESS"`
//		BaseURL         string `env:"BASE_URL"`
//		FileStoragePath string `env:"FILE_STORAGE_PATH"`
//	}
//
//	func Env() error {
//		var cfg Config
//		err := env.Parse(&cfg)
//		if err != nil {
//			fmt.Println("Невозможно спарсить cfg")
//		}
//		// Получаем значение переменной окружения
//		envVarValue, exists := os.LookupEnv("FILE_STORAGE_PATH")
//
//		if exists {
//			fmt.Printf("Значение переменной окружения %s: %s\n", "FILE_STORAGE_PATH", envVarValue)
//		} else {
//			fmt.Printf("Переменная окружения %s не установлена.\n", "FILE_STORAGE_PATH")
//		}
//		return err
//	}
//
// // Флаги
// var (
//
//	FlagServerAddress = flag.String("a", "localhost:8080", "Адрес запуска HTTP-сервера")
//	//FlagFileStoragePath = flag.String("f", "/tmp/short-url-db.json", "Полное имя файла до JSON")
//	FlagFileStoragePath = flag.String("f", "/tmp/short-url-db.json", "Полное имя файла до JSON")
//	FlagBaseURL         = flag.String("b", "http://localhost:8080", "Базовый адрес результирующего сокращённого URL")
//
// )
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
}

// Флаги
var (
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

	return &cfg, nil
}
