package rest

import (
	"net/http"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/nsec/askgod/api"
	"github.com/nsec/askgod/internal/config"
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

func (r *rest) configChanged(conf *config.Config) {
	r.eventSend("timeline", api.EventTimeline{Type: "reload"})
	return
}
