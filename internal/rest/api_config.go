package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/inconshreveable/log15"

	"github.com/nsec/askgod/api"
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

func (r *rest) updateConfig(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	// Decode the provided JSON input
	req := api.ConfigPut{}
	err := json.NewDecoder(request.Body).Decode(&req)
	if err != nil {
		logger.Warn("Malformed JSON provided", log15.Ctx{"error": err})
		r.errorResponse(400, "Malformed JSON provided", writer, request)
		return
	}

	// Save old config
	oldConfig := r.config.Config.ConfigPut
	newConfig := req

	// Attempt to update the database
	err = r.db.UpdateConfig(req)
	if err != nil {
		logger.Error("Failed to update the team", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	r.eventSend("internal", api.EventInternal{Type: "config-updated"})
	r.config.Config.ConfigPut = newConfig
	r.configHiddenTeams()
	logger.Info("Config updated", log15.Ctx{"old": oldConfig, "new": newConfig})

	// Tell everyone to reload
	r.eventSend("timeline", api.EventTimeline{Type: "reload"})
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
