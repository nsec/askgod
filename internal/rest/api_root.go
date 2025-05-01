package rest

import (
	"net/http"

	"github.com/inconshreveable/log15"
)

func (r *rest) getRoot(writer http.ResponseWriter, request *http.Request, _ log15.Logger) {
	r.jsonResponse([]string{"/1.0"}, writer, request)
}
