package logger

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/DanilCodeGit/go-yandex-shortener/internal/cfg"
	"go.uber.org/zap"
)

var sugar zap.SugaredLogger

func WithLogging(h http.Handler) http.Handler {
	// создаём предустановленный регистратор zap
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()

	// делаем регистратор SugaredLogger
	sugar = *logger.Sugar()

	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}

		// Создаем буфер для записи тела запроса
		var requestBodyBuffer bytes.Buffer
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			// Если метод POST или PUT, считываем тело запроса и записываем его в буфер
			requestBody, err := io.ReadAll(r.Body)
			if err != nil {
				// Обработайте ошибку, если не удается считать тело запроса
			} else {
				requestBodyBuffer.Write(requestBody)
				r.Body = io.NopCloser(bytes.NewBuffer(requestBody)) // Восстановите оригинальное тело запроса
			}
		}

		h.ServeHTTP(&lw, r) // внедряем реализацию http.ResponseWriter

		// Since возвращает разницу во времени между start
		// и моментом вызова Since. Таким образом можно посчитать
		// время выполнения запроса.
		duration := time.Since(start)

		// отправляем сведения о запросе в zap
		sugar.Infoln(
			"flags", *cfg.FlagFileStoragePath,
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status, // получаем перехваченный код статуса ответа
			"duration", duration,
			"size", responseData.size,
			"request_body", requestBodyBuffer.String(), // получаем перехваченный размер ответа
		)

	}
	// возвращаем функционально расширенный хендлер
	return http.HandlerFunc(logFn)
}

//func WithLogging(h http.Handler) http.Handler {
//	// создаём предустановленный регистратор zap
//	logger, err := zap.NewDevelopment()
//	if err != nil {
//		// вызываем панику, если ошибка
//		panic(err)
//	}
//	defer logger.Sync()
//
//	// делаем регистратор SugaredLogger
//	sugar := logger.Sugar()
//
//	logFn := func(w http.ResponseWriter, r *http.Request) {
//		start := time.Now()
//
//		responseData := &responseData{
//			status: 0,
//			size:   0,
//		}
//		lw := loggingResponseWriter{
//			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
//			responseData:   responseData,
//		}
//
//		// Создаем буфер для записи тела запроса
//		var requestBodyBuffer bytes.Buffer
//		if r.Method == http.MethodPost || r.Method == http.MethodPut {
//			// Если метод POST или PUT, считываем тело запроса и записываем его в буфер
//			requestBody, err := io.ReadAll(r.Body)
//			if err != nil {
//				// Обработайте ошибку, если не удается считать тело запроса
//			} else {
//				requestBodyBuffer.Write(requestBody)
//				r.Body = io.NopCloser(bytes.NewBuffer(requestBody)) // Восстановите оригинальное тело запроса
//			}
//		}
//
//		h.ServeHTTP(&lw, r) // внедряем реализацию http.ResponseWriter
//
//		// Since возвращает разницу во времени между start
//		// и моментом вызова Since. Таким образом можно посчитать
//		// время выполнения запроса.
//		duration := time.Since(start)
//
//		// отправляем сведения о запросе в zap
//		sugar.Infow(
//			"flags", *cfg.FlagFileStoragePath,
//			"uri", r.RequestURI,
//			"method", r.Method,
//			"status", responseData.status, // получаем перехваченный код статуса ответа
//			"duration", duration,
//			"size", responseData.size,
//			"request_body", requestBodyBuffer.String(), // выводим тело запроса
//		)
//	}
//
//	// возвращаем функционально расширенный хендлер
//	return http.HandlerFunc(logFn)
//}

//

type (
	// Берём структуру для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}
