package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func CreateHmacHash(r *http.Request, secret string) []byte {
	headerList := strings.Split(os.Getenv("HMAC_HEADERS"), ",")
	var hmacMessage string

	for _, v := range headerList {
		h := r.Header.Get(v)
		if os.Getenv("LOG_LEVEL") == "debug" {
			log.Printf("%v | %v", v, h)
		}
		hmacMessage = fmt.Sprintf("%v%v", hmacMessage, h)
	}
	if os.Getenv("LOG_LEVEL") == "debug" {
		log.Printf("hmac_message: %v | %v", hmacMessage, secret)
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(hmacMessage))
	hash := mac.Sum(nil)

	return hash
}
