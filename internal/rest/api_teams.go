package rest

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/inconshreveable/log15"

	"github.com/nsec/askgod/api"
)

func (r *rest) getTeam(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	// Extract the client IP
	ip, err := r.getIP(request)
	if err != nil {
		logger.Error("Failed to get the client's IP", log15.Ctx{"error": err})
		r.errorResponse(500, "Internal Server Error", writer, request)

		return
	}

	// Look for a matching team
	record, err := r.db.GetTeamForIP(*ip)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Warn("No team found for IP", log15.Ctx{"ip": ip.String()})
		r.errorResponse(404, "No team found for IP", writer, request)

		return
	} else if err != nil {
		logger.Error("Failed to get the team", log15.Ctx{"error": err})
		r.errorResponse(500, "Internal Server Error", writer, request)

		return
	}

	// Convert to the team view of a team
	team := api.Team{}
	team.ID = record.ID
	team.Name = record.Name
	team.Country = record.Country
	team.Website = record.Website

	r.jsonResponse(team, writer, request)
}

func (r *rest) updateTeam(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	// Configuration validation
	if !r.config.Teams.SelfRegister {
		logger.Warn("Unauthorized attempt to self-register")
		r.errorResponse(403, "Team self-registration disabled", writer, request)

		return
	}

	// Decode the provided JSON input
	newTeam := api.TeamPut{}
	err := json.NewDecoder(request.Body).Decode(&newTeam)
	if err != nil {
		logger.Warn("Malformed JSON provided", log15.Ctx{"error": err})
		r.errorResponse(400, "Malformed JSON provided", writer, request)

		return
	}

	// Sanity checks
	validName := func(name string) bool {
		// Validate the length
		if len(name) < 1 || len(name) > 30 {
			return false
		}

		// Validate the character set
		match, _ := regexp.MatchString("^[a-zA-Z0-9 /\\\\~!@#$%&*()\\-_+={}\\[\\];:',.?]*$", name)

		return match
	}

	if !validName(newTeam.Name) {
		logger.Warn("Bad team name", log15.Ctx{"name": newTeam.Name})
		r.errorResponse(400, "Bad team name", writer, request)

		return
	}

	match, _ := regexp.MatchString("^[A-Z]*$", newTeam.Country)
	if len(newTeam.Country) != 2 || !match {
		logger.Warn("Bad team country code", log15.Ctx{"country": newTeam.Country})
		r.errorResponse(400, "Bad team country code", writer, request)

		return
	}

	if newTeam.Website != "" {
		u, err := url.ParseRequestURI(newTeam.Website)
		if err != nil {
			logger.Warn("Bad team URL", log15.Ctx{"url": newTeam.Website})
			r.errorResponse(400, "Bad team URL", writer, request)

			return
		}

		newTeam.Website = u.String()
	}

	// Extract the client IP
	ip, err := r.getIP(request)
	if err != nil {
		logger.Error("Failed to get the client's IP", log15.Ctx{"error": err})
		r.errorResponse(500, "Internal Server Error", writer, request)

		return
	}

	// Look for a matching team
	team, err := r.db.GetTeamForIP(*ip)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Warn("No team found for IP", log15.Ctx{"ip": ip.String()})
		r.errorResponse(404, "No team found for IP", writer, request)

		return
	} else if err != nil {
		logger.Error("Failed to get the team", log15.Ctx{"error": err})
		r.errorResponse(500, "Internal Server Error", writer, request)

		return
	}

	// Validate the request
	if !r.config.Teams.SelfUpdate {
		if team.Name != "" && team.Name != newTeam.Name {
			logger.Debug("Unauthorized attempt to change already set team property", log15.Ctx{"property": "name"})
			r.errorResponse(400, "Team name is already set", writer, request)

			return
		}

		if team.Country != "" && team.Country != newTeam.Country {
			logger.Debug("Unauthorized attempt to change already set team property", log15.Ctx{"property": "country"})
			r.errorResponse(400, "Team country is already set", writer, request)

			return
		}

		if team.Website != "" && team.Website != newTeam.Website {
			logger.Debug("Unauthorized attempt to change already set team property", log15.Ctx{"property": "website"})
			r.errorResponse(400, "Team website is already set", writer, request)

			return
		}
	}

	// Setup the new record
	newRecord := api.AdminTeamPut{}
	newRecord.Name = newTeam.Name
	newRecord.Country = strings.ToUpper(newTeam.Country)
	newRecord.Website = newTeam.Website
	newRecord.Notes = team.Notes
	newRecord.Subnets = team.Subnets
	newRecord.Tags = team.Tags

	// Attempt to update the database
	err = r.db.UpdateTeam(team.ID, newRecord)
	if err != nil {
		logger.Error("Failed to update the team", log15.Ctx{"error": err})
		r.errorResponse(500, "Internal Server Error", writer, request)

		return
	}

	_ = r.eventSend("timeline", api.EventTimeline{TeamID: team.ID, Team: &newRecord.TeamPut, Type: "team-updated"})
	logger.Info("Team updated", log15.Ctx{"id": team.ID, "name": newRecord.Name, "country": newRecord.Country, "website": newRecord.Website})
}

func (r *rest) adminGetTeams(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	// Get all the teams from the database
	teams, err := r.db.GetTeams()
	if err != nil {
		logger.Error("Failed to query the team list", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)

		return
	}

	r.jsonResponse(teams, writer, request)
}

func (r *rest) adminCreateTeam(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	// Bulk create
	bulkVar := request.FormValue("bulk")
	if bulkVar == "1" {
		r.adminCreateTeams(writer, request, logger)

		return
	}

	// Decode the provided JSON input
	newTeam := api.AdminTeamPost{}
	err := json.NewDecoder(request.Body).Decode(&newTeam)
	if err != nil {
		logger.Warn("Malformed JSON provided", log15.Ctx{"error": err})
		r.errorResponse(400, "Malformed JSON provided", writer, request)

		return
	}

	// Attempt to create the database record
	id, err := r.db.CreateTeam(newTeam)
	if err != nil {
		logger.Error("Failed to create the team", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)

		return
	}

	_ = r.eventSend("timeline", api.EventTimeline{TeamID: id, Team: &newTeam.TeamPut, Type: "team-added"})
	logger.Info("New team defined", log15.Ctx{"id": id, "subnets": newTeam.Subnets})
}

func (r *rest) adminCreateTeams(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	// Decode the provided JSON input
	newTeams := []api.AdminTeamPost{}
	err := json.NewDecoder(request.Body).Decode(&newTeams)
	if err != nil {
		logger.Warn("Malformed JSON provided", log15.Ctx{"error": err})
		r.errorResponse(400, "Malformed JSON provided", writer, request)

		return
	}

	for _, team := range newTeams {
		// Attempt to create the database record
		id, err := r.db.CreateTeam(team)
		if err != nil {
			logger.Error("Failed to create the team", log15.Ctx{"error": err})
			r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)

			return
		}

		_ = r.eventSend("timeline", api.EventTimeline{TeamID: id, Team: &team.TeamPut, Type: "team-added"})
		logger.Info("New team defined", log15.Ctx{"id": id, "subnets": team.Subnets})
	}
}

func (r *rest) adminGetTeam(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	idVar := request.PathValue("id")

	// Convert the provided id to int
	id, err := strconv.ParseInt(idVar, 10, 64)
	if err != nil {
		logger.Warn("Invalid team ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(400, "Invalid team ID provided", writer, request)

		return
	}

	// Attempt to get the DB record
	team, err := r.db.GetTeam(id)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Warn("Invalid team ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(404, "Invalid team ID provided", writer, request)

		return
	} else if err != nil {
		logger.Error("Failed to get the team", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)

		return
	}

	r.jsonResponse(team, writer, request)
}

func (r *rest) adminUpdateTeam(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	idVar := request.PathValue("id")

	// Convert the provided id to int
	id, err := strconv.ParseInt(idVar, 10, 64)
	if err != nil {
		logger.Warn("Invalid team ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(400, "Invalid team ID provided", writer, request)

		return
	}

	// Decode the provided JSON input
	newTeam := api.AdminTeamPut{}
	err = json.NewDecoder(request.Body).Decode(&newTeam)
	if err != nil {
		logger.Warn("Malformed JSON provided", log15.Ctx{"error": err})
		r.errorResponse(400, "Malformed JSON provided", writer, request)

		return
	}

	// Attempt to update the database
	err = r.db.UpdateTeam(id, newTeam)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Warn("Invalid team ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(404, "Invalid team ID provided", writer, request)

		return
	} else if err != nil {
		logger.Error("Failed to update the team", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)

		return
	}

	_ = r.eventSend("timeline", api.EventTimeline{TeamID: id, Team: &newTeam.TeamPut, Type: "team-updated"})
	logger.Info("Team updated", log15.Ctx{"id": id, "name": newTeam.Name, "country": newTeam.Country, "website": newTeam.Website})
}

func (r *rest) adminDeleteTeam(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	idVar := request.PathValue("id")

	// Convert the provided id to int
	id, err := strconv.ParseInt(idVar, 10, 64)
	if err != nil {
		logger.Warn("Invalid team ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(400, "Invalid team ID provided", writer, request)

		return
	}

	// Attempt to get the DB record
	err = r.db.DeleteTeam(id)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Warn("Invalid team ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(404, "Invalid team ID provided", writer, request)

		return
	} else if err != nil {
		logger.Error("Failed to delete the team", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)

		return
	}

	_ = r.eventSend("timeline", api.EventTimeline{TeamID: id, Type: "team-removed"})
	logger.Info("Team deleted", log15.Ctx{"id": id})
}

func (r *rest) adminClearTeams(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	emptyVar := request.FormValue("empty")

	// Confirm the user is sure about it
	if emptyVar != "1" {
		logger.Warn("Teams clear requested without empty=1")
		r.errorResponse(400, "Teams clear requested without empty=1", writer, request)

		return
	}

	// Clear the database entries
	err := r.db.ClearTeams()
	if err != nil {
		logger.Error("Failed to clear all teams", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)

		return
	}

	logger.Info("All teams deleted")
}
