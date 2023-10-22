package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DanilCodeGit/go-yandex-shortener/internal/database/postgre"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/handlers"
	"github.com/magiconair/properties/assert"
)

func Test_handlePost(t *testing.T) {

	type want struct {
		statusCode  int
		responseURL string
	}
	tests := []struct {
		name        string
		requestBody string
		want        want
	}{
		{
			name:        "Empty URL",
			requestBody: "",
			want: want{
				statusCode:  http.StatusBadRequest,
				responseURL: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.requestBody)
			req, err := http.NewRequest("POST", "/", body)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handlers.HandlePost(rr, req)

			if rr.Code != tt.want.statusCode {
				t.Errorf("got status code %d, want %d", rr.Code, tt.want.statusCode)
			}

			if rr.Code == http.StatusCreated {
				responseBody := rr.Body.String()
				if responseBody != tt.want.responseURL {
					t.Errorf("got response body %s, want %s", responseBody, tt.want.responseURL)
				}
			}
		})
	}
}

func TestHandleGet(t *testing.T) {
	// Создаем запрос с необходимым путем.
	req, err := http.NewRequest("GET", "/short-url", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем записыватель ответа для захвата ответа.
	rr := httptest.NewRecorder()

	// Вызываем хэндлер.
	handlers.HandleGet(rr, req)

	// Проверяем статус ответа.
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Ожидался статус %d; получен %d", http.StatusTemporaryRedirect, rr.Code)
	}

	// Добавляем другие проверки по мере необходимости.
}

func TestHandlePost(t *testing.T) {
	// Создаем запрос с телом POST.
	postgre.DBConn(context.Background())
	requestBody := []byte("https://example.com")
	req, err := http.NewRequest("POST", "/", strings.NewReader(string(requestBody)))
	if err != nil {
		t.Fatal(err)
	}

	// Создаем записыватель ответа для захвата ответа.
	rr := httptest.NewRecorder()

	// Вызываем хэндлер.
	handlers.HandlePost(rr, req)

	// Проверяем статус ответа.
	if rr.Code != http.StatusCreated {
		t.Errorf("Ожидался статус %d; получен %d", http.StatusCreated, rr.Code)
	}

	// Проверяем тело ответа, если необходимо.
	// ...

	// Добавляем другие проверки по мере необходимости.
}

func TestJSONHandler(t *testing.T) {
	postgre.DBConn(context.Background())
	// Создаем запрос с телом JSON.
	requestBody := []byte(`{"url": "https://example.com1"}`)
	req, err := http.NewRequest("POST", "/", strings.NewReader(string(requestBody)))
	if err != nil {
		t.Fatal(err)
	}

	// Создаем записыватель ответа для захвата ответа.
	rr := httptest.NewRecorder()

	// Вызываем хэндлер.
	handlers.JSONHandler(rr, req)

	// Проверяем статус ответа.
	if rr.Code != http.StatusCreated {
		t.Errorf("Ожидался статус %d; получен %d", http.StatusCreated, rr.Code)
	}
	assert.Equal(t, rr.Code, http.StatusCreated)
	// Проверяем тело ответа, если необходимо.
	// ...

	// Добавляем другие проверки по мере необходимости.
}
