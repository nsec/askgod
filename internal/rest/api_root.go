package rest

import (
	"net/http"

	"github.com/lxc/lxd/shared/log15"
)

func (r *rest) getRoot(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	r.jsonResponse([]string{"/1.0"}, writer, request)
}
