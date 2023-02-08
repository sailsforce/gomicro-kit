package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func CreateHmacHash(r *http.Request, secret string) []byte {
	headerList := strings.Split(os.Getenv("HMAC_HEADERS"), ",")
	var hmacMessage string

	// add headers from header list
	for _, v := range headerList {
		h := r.Header.Get(v)
		if os.Getenv("LOG_LEVEL") == "debug" {
			log.Printf("%v | %v", v, h)
		}
		hmacMessage = fmt.Sprintf("%v%v", hmacMessage, h)
	}

	// add request body
	bodyBytes, _ := io.ReadAll(r.Body)
	// Restore the io.ReadCloser to its original state
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	// Use the content
	marshalReqBody, _ := json.Marshal(string(bodyBytes))
	hmacMessage = fmt.Sprintf("%v%v", hmacMessage, string(marshalReqBody))

	// add request url parameters
	hmacMessage = fmt.Sprintf("%v%v", hmacMessage, r.URL.RawQuery)

	if os.Getenv("LOG_LEVEL") == "debug" {
		log.Printf("hmac_message: %v | %v", hmacMessage, secret)
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(hmacMessage))
	hash := mac.Sum(nil)

	return hash
}
