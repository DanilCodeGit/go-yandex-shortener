package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DanilCodeGit/go-yandex-shortener/internal/handlers"
	"github.com/DanilCodeGit/go-yandex-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
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

func Test_handleGet(t *testing.T) {
	storage.URLStore = map[string]string{
		"abc123": "http://example.com",
		"def456": "http://example.org",
	}

	type want struct {
		statusCode  int
		originalURL string
	}
	tests := []struct {
		name       string
		requestURL string
		want       want
	}{
		{
			name:       "Valid URL",
			requestURL: "/abc123",
			want: want{
				statusCode:  http.StatusTemporaryRedirect,
				originalURL: "http://example.com",
			},
		},
		{
			name:       "Nonexistent URL",
			requestURL: "/nonexistent",
			want: want{
				statusCode:  http.StatusTemporaryRedirect,
				originalURL: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.requestURL, nil)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			handlers.HandleGet(w, req)
			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, res.StatusCode, tt.want.statusCode)
			if assert.Equal(t, res.StatusCode, http.StatusTemporaryRedirect) {
				locationHeader := w.Header().Get("Location")
				assert.Equal(t, locationHeader, tt.want.originalURL)

			}
		})
	}
}
