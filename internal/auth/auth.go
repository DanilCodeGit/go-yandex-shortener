package auth

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Claims — структура утверждений, которая включает стандартные утверждения
// и одно пользовательское — UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

const TOKEN_EXP = time.Hour * 3
const SECRET_KEY = "supersecretkey"

// BuildJWTString создаёт токен и возвращает его в виде строки.
func BuildJWTString(id int) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		// собственное утверждение
		UserID: id,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}
func GetUserId(tokenString string) int {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		})
	if err != nil {
		return -1
	}

	if !token.Valid {
		fmt.Println("Token is not valid")
		return -1
	}

	fmt.Println("Token is valid")
	return claims.UserID
}

func GenerateRandomID() (int, error) {
	// Генерируем случайное байтовое значение
	randomBytes := make([]byte, len(SECRET_KEY))
	_, err := rand.Read(randomBytes)
	if err != nil {
		return 0, err
	}

	// Преобразуем байтовое значение в целое число
	id := int(binary.BigEndian.Uint64(randomBytes))
	return id, nil
}

func AuthMiddleWare(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Попытка получить куку JWT
		cookie, err := r.Cookie("jwt")
		if err != nil || cookie.Value == "" {
			id, err := GenerateRandomID()
			if err != nil {
				http.Error(w, "Не удалось сгенерировать случайный ID", http.StatusInternalServerError)
				return
			}
			// Создание JWT-токена
			tokenString, err := BuildJWTString(id)
			if err != nil {
				http.Error(w, "Не удалось создать JWT-токен", http.StatusInternalServerError)
				return
			}

			// Сохранение JWT-токена в куке
			http.SetCookie(w, &http.Cookie{
				Name:  "jwt",
				Value: tokenString,
				//Expires:  time.Now().Add(TOKEN_EXP),
				HttpOnly: true,
			})
			// Вызов обернутого обработчика с токеном в куке
			h(w, r)
		} else {
			// Если кука существует, попытка извлечь ID пользователя из токена
			userID := GetUserId(cookie.Value)
			if userID == -1 {
				http.Error(w, "Недействительный JWT-токен", http.StatusUnauthorized)
				return
			}
			// Вызов обернутого обработчика с извлеченным ID пользователя
			// Вы должны передавать userID в обработчик или использовать его по необходимости
			h(w, r)
		}
	}
}
