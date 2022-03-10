package utils

import (
	"net/http"
)

func CopyRequestHeaders(r *http.Request, target *http.Request) {
	for k, vv := range r.Header {
		for _, v := range vv {
			target.Header.Set(k, v)
		}
	}
}
