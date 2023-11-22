package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DanilCodeGit/go-yandex-shortener/internal/cfg"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/database/postgre"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/handlers"
	"github.com/magiconair/properties/assert"
)

func TestHandleGet(t *testing.T) {
	// Создаем запрос с необходимым путем.
	conn, _ := postgre.NewDataBase(context.Background(), *cfg.FlagDataBaseDSN)
	req, err := http.NewRequest("GET", "/short-url", nil)
	if err != nil {
		t.Fatal(err)
	}
	dbHandle := handlers.DataBaseHandle{DB: conn}
	// Создаем записыватель ответа для захвата ответа.
	rr := httptest.NewRecorder()
	// Вызываем хэндлер.
	dbHandle.HandleGet().ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusTemporaryRedirect)

	// Добавляем другие проверки по мере необходимости.
}
