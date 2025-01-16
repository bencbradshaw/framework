package middleware

import (
	"fmt"
	"net/http"
	"time"
)

// LoggingMiddleware logs the incoming HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		fmt.Printf("START: [%s %s]\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		finishTime := time.Now()
		totalTime := finishTime.Sub(startTime)
		fmt.Printf("[%s %d %v] %s\n", r.Method, http.StatusOK, totalTime, r.URL.Path)
	})
}
