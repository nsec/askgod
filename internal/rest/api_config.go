package rest

import (
	"net/http"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/nsec/askgod/api"
	"github.com/nsec/askgod/internal/config"
	"github.com/nsec/askgod/internal/utils"
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

func (r *rest) configHiddenTeams() error {
	teamIDs := []int64{}
	teams, err := r.db.GetTeams()
	if err != nil {
		r.logger.Error("Unable to refresh hidden teams", log15.Ctx{"error": err})
		return err
	}

	for _, team := range teams {
		if utils.StringInSlice(team.Name, r.config.Teams.Hidden) {
			teamIDs = append(teamIDs, team.ID)
		}
	}
	r.hiddenTeams = teamIDs

	return nil
}

func (r *rest) configChanged(conf *config.Config) {
	// Update the list of hidden teams
	r.configHiddenTeams()

	// Tell everyone to reload
	r.eventSend("timeline", api.EventTimeline{Type: "reload"})
	return
}
