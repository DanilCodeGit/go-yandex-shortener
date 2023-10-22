package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DanilCodeGit/go-yandex-shortener/internal/cfg"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/database/postgre"
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

	// Проверяем статус ответа.
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Ожидался статус %d; получен %d", http.StatusTemporaryRedirect, rr.Code)
	}

	// Добавляем другие проверки по мере необходимости.
}

func TestHandlePost(t *testing.T) {
	// Создаем запрос с телом POST.
	base, err := postgre.NewDataBase(context.Background(), *cfg.FlagDataBaseDSN)
	if err != nil {
		return
	}
	defer base.Close()
	requestBody := []byte("https://example.com")
	_, err = http.NewRequest("POST", "/", strings.NewReader(string(requestBody)))
	if err != nil {
		t.Fatal(err)
	}

	// Создаем записыватель ответа для захвата ответа.
	rr := httptest.NewRecorder()

	// Вызываем хэндлер.
	handlers.HandlePost(base)

	// Проверяем статус ответа.
	if rr.Code != http.StatusCreated {
		t.Errorf("Ожидался статус %d; получен %d", http.StatusCreated, rr.Code)
	}

	// Проверяем тело ответа, если необходимо.
	// ...

	// Добавляем другие проверки по мере необходимости.
}

func TestJSONHandler(t *testing.T) {
	base, err := postgre.NewDataBase(context.Background(), *cfg.FlagDataBaseDSN)
	if err != nil {
		return
	}
	defer base.Close()

	requestBody := []byte(`{"url": "https://example.com4"}`)
	_, err = http.NewRequest("POST", "/", strings.NewReader(string(requestBody)))
	if err != nil {
		t.Fatal(err)
	}

	// Создаем записыватель ответа для захвата ответа.
	rr := httptest.NewRecorder()

	// Вызываем хэндлер.
	handlers.JSONHandler(base)

	// Проверяем статус ответа.
	if rr.Code != http.StatusCreated {
		t.Errorf("Ожидался статус %d; получен %d", http.StatusCreated, rr.Code)
	}
	assert.Equal(t, rr.Code, http.StatusCreated)
	// Проверяем тело ответа, если необходимо.
	// ...

	// Добавляем другие проверки по мере необходимости.
}
