package middleware

import (
	"log"
	"net/http"
)

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		log.Println("adding cors header")

		rw.Header().Set("Access-Control-Allow-Origin", "*")

		next.ServeHTTP(rw, r)
	})
}
