package middleware

import (
	"bytes"
	"crypto/hmac"
	"encoding/base64"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	kit_errors "github.com/sailsforce/gomicro-kit/errors"
	kit_logger "github.com/sailsforce/gomicro-kit/logger"
	kit_models "github.com/sailsforce/gomicro-kit/models"
	kit_utils "github.com/sailsforce/gomicro-kit/utils"
)

func HmacHash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		logger := kit_logger.GetLogEntry(r)
		reqId := middleware.GetReqID(r.Context())

		logger.Info("creating hmach hash...")

		keys, err := loadHmacKeys()
		if err != nil {
			logger.Error("error loading hmac keys: ", err)
			logger.Debug(render.Render(rw, r, kit_errors.ErrInternal(reqId)))
			return
		}

		key := keys.GetLatestKey()
		logger.Info("retrieved latest key.")
		logger.Debug("key: ", key)

		hmacByte := kit_utils.CreateHmacHash(r, key)
		logger.Debug("hmac hash: ", hmacByte)

		//base64 encode
		hmac64 := base64.StdEncoding.EncodeToString(hmacByte)
		// add to headers
		r.Header.Add("X-HMAC-HASH", hmac64)
		logger.Info("hmac added to header.")

		next.ServeHTTP(rw, r)
	})
}

func ValidateHmac(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		logger := kit_logger.GetLogEntry(r)
		reqId := middleware.GetReqID(r.Context())

		logger.Info("validating hmac...")

		logger.Debug("getting hmac from header...")
		headerValue64 := r.Header.Get("X-HMAC-HASH")
		logger.Debug("retrieved from header.")

		logger.Debug("create hash for validation.")
		keys, err := loadHmacKeys()
		if err != nil {
			logger.Error("error loading hmac keys: ", err)
			logger.Debug(render.Render(rw, r, kit_errors.ErrInternal(reqId)))
			return
		}

		// validate
		headerValue, err := base64.StdEncoding.DecodeString(headerValue64)
		if err != nil {
			logger.Error("error decoding hmac hash from header: ", err)
			logger.Debug(render.Render(rw, r, kit_errors.ErrInvalidRequest(reqId)))
			return
		}
		logger.Debug("header hmac: ", headerValue)

		logger.Info("validating...")
		validated := validateHmacKeys(keys, headerValue, r)
		if !validated {
			logger.Error("hmac did not match. Forbidden request")
			logger.Debug(render.Render(rw, r, kit_errors.ErrFobiddenRequest(reqId)))
			return
		}

		next.ServeHTTP(rw, r)
	})
}

func validateHmacKeys(keys *kit_models.HmacKeys, headerHmac []byte, req *http.Request) bool {
	logger := kit_logger.GetLogEntry(req)
	for _, v := range keys.Keys {
		expected := kit_utils.CreateHmacHash(req, v.Value)
		logger.Debug("Checking header against: ", expected)
		if hmac.Equal(headerHmac, expected) {
			return true
		}
	}
	logger.Info("no keys matched.")
	return false
}

func loadHmacKeys() (*kit_models.HmacKeys, error) {
	var keys *kit_models.HmacKeys
	err := render.DecodeJSON(bytes.NewReader([]byte(os.Getenv("HMAC_SECRETS"))), &keys)
	if err != nil {
		return nil, err
	}

	return keys, nil
}
