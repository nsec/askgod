package rest

import (
	"net/http"

	"github.com/inconshreveable/log15"

	"github.com/nsec/askgod/api"
)

func (r *rest) getStatus(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	resp := api.Status{
		IsAdmin:   r.hasAccess("admin", request),
		IsTeam:    r.hasAccess("team", request),
		IsGuest:   r.hasAccess("guest", request),
		EventName: r.config.Scoring.EventName,
		Flags: api.StatusFlags{
			TeamSelfRegister: r.config.Teams.SelfRegister,
			TeamSelfUpdate:   r.config.Teams.SelfUpdate,
			BoardReadOnly:    r.config.Scoring.ReadOnly,
			BoardHideOthers:  r.config.Scoring.HideOthers,
		},
	}

	r.jsonResponse(resp, writer, request)
}
