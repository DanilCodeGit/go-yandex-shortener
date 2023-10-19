package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DanilCodeGit/go-yandex-shortener/internal/handlers"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/storage"
	"github.com/magiconair/properties/assert"
)

func Test_handlePost(t *testing.T) {
	storage.URLStore = map[string]string{}
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

//func Test_handleGet(t *testing.T) {
//	storage.URLStore = map[string]string{
//		"abc123": "http://example.com",
//		"def456": "http://example.org",
//	}
//
//	type want struct {
//		statusCode  int
//		originalURL string
//	}
//	tests := []struct {
//		name       string
//		requestURL string
//		want       want
//	}{
//		{
//			name:       "Valid URL",
//			requestURL: "/abc123",
//			want: want{
//				statusCode:  http.StatusTemporaryRedirect,
//				originalURL: "http://example.com",
//			},
//		},
//		{
//			name:       "Nonexistent URL",
//			requestURL: "/nonexistent",
//			want: want{
//				statusCode:  http.StatusTemporaryRedirect,
//				originalURL: "",
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			req, err := http.NewRequest("GET", tt.requestURL, nil)
//			if err != nil {
//				t.Fatal(err)
//			}
//			w := httptest.NewRecorder()
//			handlers.HandleGet(w, req)
//			res := w.Result()
//			defer res.Body.Close()
//			assert.Equal(t, res.StatusCode, tt.want.statusCode)
//			if assert.Equal(t, res.StatusCode, http.StatusTemporaryRedirect) {
//				locationHeader := w.Header().Get("Location")
//				assert.Equal(t, locationHeader, tt.want.originalURL)
//
//			}
//		})
//	}
//}

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
