package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DanilCodeGit/go-yandex-shortener/internal/handlers"
	"github.com/magiconair/properties/assert"
)

func TestHandleGet(t *testing.T) {
	// Создаем запрос с необходимым путем.
	req, err := http.NewRequest("GET", "/short-url", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем записыватель ответа для захвата ответа.
	rr := httptest.NewRecorder()
	// Вызываем хэндлер.
	handlers.HandleGet().ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusTemporaryRedirect)

	// Добавляем другие проверки по мере необходимости.
}

//func TestHandlePost(t *testing.T) {
//	// Создаем запрос с телом POST.
//	conn, err := postgre.NewDataBase(context.Background(), *cfg.FlagDataBaseDSN)
//	if err != nil {
//		log.Fatal("Database connection failed")
//	}
//	requestBody := []byte("https://example.com54")
//	wantBody := tools.HashURL(string(requestBody))
//	_, err = http.NewRequest("POST", "/", strings.NewReader(string(requestBody)))
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// Создаем записыватель ответа для захвата ответа.
//	rr := httptest.NewRecorder()
//
//	// Вызываем хэндлер.
//	handlers.HandlePost(conn)
//
//	assert.Equal(t, rr.Code, http.StatusCreated)
//	assert.Equal(t, rr.Body, wantBody)
//	// Проверяем тело ответа, если необходимо.
//	// ...
//
//	// Добавляем другие проверки по мере необходимости.
//}
//
//func TestJSONHandler(t *testing.T) {
//	base, err := postgre.NewDataBase(context.Background(), *cfg.FlagDataBaseDSN)
//	if err != nil {
//		return
//	}
//
//	requestBody := []byte(`{"url": "https://example.com4"}`)
//	_, err = http.NewRequest("POST", "/", strings.NewReader(string(requestBody)))
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// Создаем записыватель ответа для захвата ответа.
//	rr := httptest.NewRecorder()
//
//	// Вызываем хэндлер.
//	handlers.JSONHandler(base)
//
//	// Проверяем статус ответа.
//	if rr.Code != http.StatusCreated {
//		t.Errorf("Ожидался статус %d; получен %d", http.StatusCreated, rr.Code)
//	}
//	assert.Equal(t, rr.Code, http.StatusCreated)
//	// Проверяем тело ответа, если необходимо.
//	// ...
//
//	// Добавляем другие проверки по мере необходимости.
//}
