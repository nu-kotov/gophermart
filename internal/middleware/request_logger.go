package middleware

import (
	"net/http"
	"time"

	"github.com/nu-kotov/gophermart/internal/logger"
	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func RequestLogger(h http.HandlerFunc) http.HandlerFunc {
	logFn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(&lw, r)
		duration := time.Since(start)
		logger.Log.Info("request",
			zap.String("method", r.Method),
			zap.String("uri", r.URL.Path),
			zap.String("duration", duration.String()),
		)
		logger.Log.Info("response",
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
		)
	})
	return http.HandlerFunc(logFn)
}
