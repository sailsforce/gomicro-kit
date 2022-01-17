package middleware

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

func Headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		requestId := middleware.GetReqID(r.Context())
		log.Println("adding default headers")

		rw.Header().Add("X-Frame-Options", "DENY")
		rw.Header().Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		rw.Header().Add("X-SCI-REQUESTID", requestId)

		next.ServeHTTP(rw, r)
	})
}
