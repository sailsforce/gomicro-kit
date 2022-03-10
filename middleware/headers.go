package middleware

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5/middleware"
)

func Headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		requestId := middleware.GetReqID(r.Context())
		log.Println("adding default headers")

		header := os.Getenv("REQ_ID_HEADER")
		if header == "" {
			header = "X-SERVICE-REQUESTID"
		}

		rw.Header().Add("X-Frame-Options", "DENY")
		rw.Header().Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		rw.Header().Add(header, requestId)

		next.ServeHTTP(rw, r)
	})
}
