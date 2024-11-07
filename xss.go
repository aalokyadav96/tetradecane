package main

// import (
// 	"net/http"
// )

// // Security headers middleware
// func securityHeaders(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("X-XSS-Protection", "1; mode=block")
// 		w.Header().Set("X-Content-Type-Options", "nosniff")
// 		w.Header().Set("X-Frame-Options", "DENY")
// 		next.ServeHTTP(w, r) // Call the next handler
// 	})
// }

// func main() {
// 	http.Handle("/", securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte("Hello, World! This page is secure."))
// 	})))

// 	http.ListenAndServe(":8080", nil)
// }
