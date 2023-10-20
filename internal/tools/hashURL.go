package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"os"
)

func HashURL(urlToHash string) string {
	// Создайте новый хеш SHA-256
	hasher := sha256.New()

	// Преобразуйте URL в байтовый срез и передайте его хешеру
	hasher.Write([]byte(urlToHash))

	// Получите байтовое представление хеша
	hashBytes := hasher.Sum(nil)

	// Преобразуйте байты хеша в строку в шестнадцатеричном формате
	hashedURL := hex.EncodeToString(hashBytes)
	// Ограничиваем строку первыми 5 символами
	runes := []rune(hashedURL)
	shortenedHashedURL := string(runes[:6])

	return shortenedHashedURL
}

func IsFlagAndEnvSet(flagName string, envName string) bool {
	flagSet := false
	envSet := false

	// Проверяем, установлен ли флаг
	flag.Visit(func(f *flag.Flag) {
		if f.Name == flagName {
			flagSet = true
		}
	})

	// Проверяем, установлена ли переменная окружения
	if envValue, exists := os.LookupEnv(envName); exists {
		if envValue != "" {
			envSet = true
		}
	}

	// Возвращаем true, если и флаг, и переменная окружения установлены
	return flagSet && envSet
}
