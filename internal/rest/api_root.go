package rest

import (
	"net/http"

	"gopkg.in/inconshreveable/log15.v2"
)

func (r *rest) getRoot(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	r.jsonResponse([]string{"/1.0"}, writer, request)
}
