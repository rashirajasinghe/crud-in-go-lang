package router

import (
	"net/http"
	"time"

	"crud-in-go-lang/internal/controller"

	"github.com/gorilla/mux"
)


func SetupRouter(bookController *controller.BookController) *mux.Router {
	r := mux.NewRouter()


	r.Use(loggingMiddleware)
	r.Use(corsMiddleware)


	bookController.RegisterRoutes(r)


	r.HandleFunc("/health", healthCheckHandler).Methods("GET")

	return r
}


func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()


		next.ServeHTTP(w, r)


		duration := time.Since(start)
		requestLog := map[string]interface{}{
			"method":     r.Method,
			"path":       r.URL.Path,
			"query":      r.URL.RawQuery,
			"duration":   duration.String(),
			"user_agent": r.UserAgent(),
			"remote_addr": r.RemoteAddr,
		}


		_ = requestLog
	})
}


func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")


		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}


		next.ServeHTTP(w, r)
	})
}


func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}