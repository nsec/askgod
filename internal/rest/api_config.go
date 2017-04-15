package rest

import (
	"net/http"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/nsec/askgod/api"
)

func (r *rest) getConfig(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	resp := api.Config(*r.config.Config)
	if resp.Daemon.HTTPSCertificate != "" {
		resp.Daemon.HTTPSCertificate = "*****"
	}

	if resp.Daemon.HTTPSKey != "" {
		resp.Daemon.HTTPSKey = "*****"
	}

	if resp.Database.Password != "" {
		resp.Database.Password = "*****"
	}

	r.jsonResponse(resp, writer, request)
}
