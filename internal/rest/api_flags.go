package rest

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"

	"github.com/nsec/askgod/api"
)

func (r *rest) getTeamFlags(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	// Extract the client IP
	ip, err := r.getIP(request)
	if err != nil {
		logger.Error("Failed to get the client's IP", log15.Ctx{"error": err})
		r.errorResponse(500, "Internal Server Error", writer, request)
		return
	}

	// Look for a matching team
	team, err := r.db.GetTeamForIP(*ip)
	if err == sql.ErrNoRows {
		logger.Warn("No team found for IP", log15.Ctx{"ip": ip.String()})
		r.errorResponse(404, "No team found for IP", writer, request)
		return
	} else if err != nil {
		logger.Error("Failed to get the team", log15.Ctx{"error": err})
		r.errorResponse(500, "Internal Server Error", writer, request)
		return
	}

	// Get all the flags for the team
	flags, err := r.db.GetTeamFlags(team.ID)
	if err != nil {
		logger.Error("Failed to query the flag list", log15.Ctx{"error": err, "teamid": team.ID})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	r.jsonResponse(flags, writer, request)
}

func (r *rest) getTeamFlag(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	idVar := mux.Vars(request)["id"]

	// Convert the provided id to int
	id, err := strconv.ParseInt(idVar, 10, 64)
	if err != nil {
		logger.Warn("Invalid flag ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(400, "Invalid flag ID provided", writer, request)
		return
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
	if err == sql.ErrNoRows {
		logger.Warn("No team found for IP", log15.Ctx{"ip": ip.String()})
		r.errorResponse(404, "No team found for IP", writer, request)
		return
	} else if err != nil {
		logger.Error("Failed to get the team", log15.Ctx{"error": err})
		r.errorResponse(500, "Internal Server Error", writer, request)
		return
	}

	// Get all the flags for the team
	flag, err := r.db.GetTeamFlag(team.ID, id)
	if err != nil {
		logger.Error("Failed to query the flag", log15.Ctx{"error": err, "teamid": team.ID, "flagid": id})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	r.jsonResponse(flag, writer, request)
}

func (r *rest) updateTeamFlag(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	idVar := mux.Vars(request)["id"]

	// Convert the provided id to int
	id, err := strconv.ParseInt(idVar, 10, 64)
	if err != nil {
		logger.Warn("Invalid flag ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(400, "Invalid flag ID provided", writer, request)
		return
	}

	// Decode the provided JSON input
	flag := api.FlagPut{}
	err = json.NewDecoder(request.Body).Decode(&flag)
	if err != nil {
		logger.Warn("Malformed JSON provided", log15.Ctx{"error": err})
		r.errorResponse(400, "Malformed JSON provided", writer, request)
		return
	}

	// Validate the input
	if len(flag.Notes) > 1000 {
		logger.Warn("Note is too long")
		r.errorResponse(400, "Note is too long", writer, request)
		return
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
	if err == sql.ErrNoRows {
		logger.Warn("No team found for IP", log15.Ctx{"ip": ip.String()})
		r.errorResponse(404, "No team found for IP", writer, request)
		return
	} else if err != nil {
		logger.Error("Failed to get the team", log15.Ctx{"error": err})
		r.errorResponse(500, "Internal Server Error", writer, request)
		return
	}

	// Update the team flag
	err = r.db.UpdateTeamFlag(team.ID, id, flag)
	if err != nil {
		logger.Error("Failed to update the flag", log15.Ctx{"error": err, "teamid": team.ID, "flagid": id})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}
}

func (r *rest) submitTeamFlag(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	// Check if read-only
	if r.config.Scoring.ReadOnly {
		r.errorResponse(403, "Flag submission isn't allowed at this time", writer, request)
		return
	}

	// Decode the provided JSON input
	flag := api.FlagPost{}
	err := json.NewDecoder(request.Body).Decode(&flag)
	if err != nil {
		logger.Warn("Malformed JSON provided", log15.Ctx{"error": err})
		r.errorResponse(400, "Malformed JSON provided", writer, request)
		return
	}

	// Validate the input
	if len(flag.Notes) > 1000 {
		logger.Warn("Note is too long")
		r.errorResponse(400, "Note is too long", writer, request)
		return
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
	if err == sql.ErrNoRows {
		logger.Warn("No team found for IP", log15.Ctx{"ip": ip.String()})
		r.errorResponse(404, "No team found for IP", writer, request)
		return
	} else if err != nil {
		logger.Error("Failed to get the team", log15.Ctx{"error": err})
		r.errorResponse(500, "Internal Server Error", writer, request)
		return
	}

	// Check that the team is configured
	if team.Name == "" || team.Country == "" {
		logger.Debug("Unconfigured team tried to submit flag", log15.Ctx{"teamid": team.ID})
		r.errorResponse(400, "Team name and country are required to participate", writer, request)
		return
	}

	// Submit the flag
	result, adminFlag, err := r.db.SubmitTeamFlag(team.ID, flag)
	if err == sql.ErrNoRows {
		eventSend("flags", api.EventFlag{Team: *team, Input: flag.Flag, Type: "invalid"})
		logger.Info("Invalid flag submitted", log15.Ctx{"teamid": team.ID, "flag": flag.Flag})
		r.errorResponse(400, "Invalid flag submitted", writer, request)
		return
	} else if err == os.ErrExist {
		eventSend("flags", api.EventFlag{Team: *team, Flag: adminFlag, Input: flag.Flag, Value: 0, Type: "duplicate"})
		logger.Info("The flag was already submitted", log15.Ctx{"teamid": team.ID, "flag": flag.Flag})
		r.errorResponse(400, "The flag was already submitted", writer, request)
		return
	} else if err != nil {
		logger.Error("Failed to submit the flag", log15.Ctx{"error": err, "teamid": team.ID})
		r.errorResponse(500, "Internal Server Error", writer, request)
		return
	}

	// Send the flag notification
	eventSend("flags", api.EventFlag{Team: *team, Flag: adminFlag, Input: flag.Flag, Value: result.Value, Type: "valid"})

	// Send the timeline notification
	total, err := r.db.GetTeamPoints(team.ID)
	if err != nil {
		logger.Error("Failed to get the team score record", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	score := api.TimelineEntryScore{
		SubmitTime: time.Now(),
		Value:      result.Value,
		Total:      total,
	}

	eventSend("timeline", api.EventTimeline{TeamID: team.ID, Team: &team.AdminTeamPut.TeamPut, Score: &score, Type: "score-updated"})

	logger.Info("Correct flag submitted", log15.Ctx{"teamid": team.ID, "flagid": result.ID, "value": result.Value, "flag": flag.Flag})
	r.jsonResponse(result, writer, request)
}

func (r *rest) adminGetFlags(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	// Get all the flags from the database
	flags, err := r.db.GetFlags()
	if err != nil {
		logger.Error("Failed to query the flag list", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	r.jsonResponse(flags, writer, request)
}

func (r *rest) adminCreateFlag(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	// Bulk create
	bulkVar := request.FormValue("bulk")
	if bulkVar == "1" {
		r.adminCreateFlags(writer, request, logger)
		return
	}

	// Decode the provided JSON input
	newFlag := api.AdminFlagPost{}
	err := json.NewDecoder(request.Body).Decode(&newFlag)
	if err != nil {
		logger.Warn("Malformed JSON provided", log15.Ctx{"error": err})
		r.errorResponse(400, "Malformed JSON provided", writer, request)
		return
	}

	// Attempt to update the database
	id, err := r.db.CreateFlag(newFlag)
	if err != nil {
		logger.Error("Failed to create the flag", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	logger.Info("New flag defined", log15.Ctx{"id": id, "flag": newFlag.Flag, "value": newFlag.Value})
}

func (r *rest) adminCreateFlags(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	// Decode the provided JSON input
	newFlags := []api.AdminFlagPost{}
	err := json.NewDecoder(request.Body).Decode(&newFlags)
	if err != nil {
		logger.Warn("Malformed JSON provided", log15.Ctx{"error": err})
		r.errorResponse(400, "Malformed JSON provided", writer, request)
		return
	}

	for _, flag := range newFlags {
		// Attempt to create the database record
		id, err := r.db.CreateFlag(flag)
		if err != nil {
			logger.Error("Failed to create the flag", log15.Ctx{"error": err})
			r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
			return
		}

		logger.Info("New flag defined", log15.Ctx{"id": id, "flag": flag.Flag, "value": flag.Value})
	}
}

func (r *rest) adminGetFlag(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	idVar := mux.Vars(request)["id"]

	// Convert the provided id to int
	id, err := strconv.ParseInt(idVar, 10, 64)
	if err != nil {
		logger.Warn("Invalid flag ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(400, "Invalid flag ID provided", writer, request)
		return
	}

	// Attempt to get the DB record
	flag, err := r.db.GetFlag(id)
	if err == sql.ErrNoRows {
		logger.Warn("Invalid flag ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(404, "Invalid flag ID provided", writer, request)
		return
	} else if err != nil {
		logger.Error("Failed to get the flag", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	r.jsonResponse(flag, writer, request)
}

func (r *rest) adminUpdateFlag(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	idVar := mux.Vars(request)["id"]

	// Convert the provided id to int
	id, err := strconv.ParseInt(idVar, 10, 64)
	if err != nil {
		logger.Warn("Invalid flag ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(400, "Invalid flag ID provided", writer, request)
		return
	}

	// Decode the provided JSON input
	newFlag := api.AdminFlagPut{}
	err = json.NewDecoder(request.Body).Decode(&newFlag)
	if err != nil {
		logger.Warn("Malformed JSON provided", log15.Ctx{"error": err})
		r.errorResponse(400, "Malformed JSON provided", writer, request)
		return
	}

	// Attempt to update the database
	err = r.db.UpdateFlag(id, newFlag)
	if err == sql.ErrNoRows {
		logger.Warn("Invalid flag ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(404, "Invalid flag ID provided", writer, request)
		return
	} else if err != nil {
		logger.Error("Failed to update the flag", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	logger.Info("Flag updated", log15.Ctx{"id": id, "flag": newFlag.Flag, "value": newFlag.Value})
}

func (r *rest) adminDeleteFlag(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	idVar := mux.Vars(request)["id"]

	// Convert the provided id to int
	id, err := strconv.ParseInt(idVar, 10, 64)
	if err != nil {
		logger.Warn("Invalid flag ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(400, "Invalid flag ID provided", writer, request)
		return
	}

	// Attempt to get the DB record
	err = r.db.DeleteFlag(id)
	if err == sql.ErrNoRows {
		logger.Warn("Invalid flag ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(404, "Invalid flag ID provided", writer, request)
		return
	} else if err != nil {
		logger.Error("Failed to delete the flag", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	logger.Info("Flag deleted", log15.Ctx{"id": id})
}

func (r *rest) adminClearFlags(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	emptyVar := request.FormValue("empty")

	// Confirm the user is sure about it
	if emptyVar != "1" {
		logger.Warn("Flags clear requested without empty=1")
		r.errorResponse(400, "Flags clear requested without empty=1", writer, request)
		return
	}

	// Clear the database entries
	err := r.db.ClearFlags()
	if err != nil {
		logger.Error("Failed to clear all flags", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	logger.Info("All flags deleted")
}
