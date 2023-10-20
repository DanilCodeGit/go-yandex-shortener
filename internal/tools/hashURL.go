package tools

import (
	"crypto/sha256"
	"encoding/hex"
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

func IsFlagAndEnvSet(envName string) bool {

	envSet := false

	// Проверяем, установлена ли переменная окружения
	if envValue, exists := os.LookupEnv(envName); exists {
		if envValue != "" {
			envSet = true
		}
	}

	// Возвращаем true, если и флаг, и переменная окружения установлены
	return envSet
}
