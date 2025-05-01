package rest

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/inconshreveable/log15"

	"github.com/nsec/askgod/api"
	"github.com/nsec/askgod/internal/utils"
)

func (r *rest) getScoreboard(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	// If scoreboard hidden and not a team, show empty board
	if r.config.Scoring.HideOthers && !r.hasAccess("team", request) {
		r.jsonResponse([]api.ScoreboardEntry{}, writer, request)

		return
	}

	// Get the full scoreboard
	scoreboard, err := r.db.GetScoreboard()
	if err != nil {
		logger.Error("Failed to get the scoreboard", log15.Ctx{"error": err})
		r.errorResponse(500, "Internal Server Error", writer, request)

		return
	}

	// Filter the results
	if (r.config.Scoring.HideOthers || len(r.hiddenTeams) > 0) && !r.hasAccess("admin", request) {
		// Extract the client IP
		ip, err := r.getIP(request)
		if err != nil {
			logger.Error("Failed to get the client's IP", log15.Ctx{"error": err})
			r.errorResponse(500, "Internal Server Error", writer, request)

			return
		}

		// Look for a matching team
		var team *api.AdminTeam
		if r.hasAccess("team", request) {
			team, err = r.db.GetTeamForIP(*ip)
			if errors.Is(err, sql.ErrNoRows) {
				logger.Warn("No team found for IP", log15.Ctx{"ip": ip.String()})
				r.errorResponse(404, "No team found for IP", writer, request)

				return
			} else if err != nil {
				logger.Error("Failed to get the team", log15.Ctx{"error": err})
				r.errorResponse(500, "Internal Server Error", writer, request)

				return
			}
		}

		newBoard := []api.ScoreboardEntry{}
		for _, entry := range scoreboard {
			if r.config.Scoring.HideOthers && (team == nil || entry.Team.ID != team.ID) {
				continue
			}

			if utils.Int64InSlice(entry.Team.ID, r.hiddenTeams) && (team == nil || team.ID != entry.Team.ID) {
				continue
			}

			newBoard = append(newBoard, entry)
		}

		scoreboard = newBoard
	}

	r.jsonResponse(scoreboard, writer, request)
}
