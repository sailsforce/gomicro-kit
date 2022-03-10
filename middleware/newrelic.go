package middleware

import (
	"net/http"

	"github.com/newrelic/go-agent/v3/newrelic"
)

func NewRelicWrapper(next http.Handler, newRelicApp *newrelic.Application) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if newRelicApp != nil {
			txn := newRelicApp.StartTransaction(r.Method + " " + r.RequestURI)
			defer txn.End()
			txn.SetWebRequestHTTP(r)
			rw = txn.SetWebResponse(rw)
			r = newrelic.RequestWithTransactionContext(r, txn)
		}

		next.ServeHTTP(rw, r)
	})
}
