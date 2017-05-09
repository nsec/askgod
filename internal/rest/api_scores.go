package rest

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"

	"github.com/nsec/askgod/api"
)

func (r *rest) adminGetScores(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	// Get all the scores from the database
	scores, err := r.db.GetScores()
	if err != nil {
		logger.Error("Failed to query the score list", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	r.jsonResponse(scores, writer, request)
}

func (r *rest) adminCreateScore(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	// Bulk create
	bulkVar := request.FormValue("bulk")
	if bulkVar == "1" {
		r.adminCreateScores(writer, request, logger)
		return
	}

	// Decode the provided JSON input
	newScore := api.AdminScorePost{}
	err := json.NewDecoder(request.Body).Decode(&newScore)
	if err != nil {
		logger.Warn("Malformed JSON provided", log15.Ctx{"error": err})
		r.errorResponse(400, "Malformed JSON provided", writer, request)
		return
	}

	r.adminCreateScoreCommon(writer, request, logger, newScore)
}

func (r *rest) adminCreateScores(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	// Decode the provided JSON input
	newScores := []api.AdminScorePost{}
	err := json.NewDecoder(request.Body).Decode(&newScores)
	if err != nil {
		logger.Warn("Malformed JSON provided", log15.Ctx{"error": err})
		r.errorResponse(400, "Malformed JSON provided", writer, request)
		return
	}

	for _, score := range newScores {
		if !r.adminCreateScoreCommon(writer, request, logger, score) {
			return
		}
	}
}

func (r *rest) adminCreateScoreCommon(writer http.ResponseWriter, request *http.Request, logger log15.Logger, newScore api.AdminScorePost) bool {
	// Attempt to update the database
	id, err := r.db.CreateScore(newScore)
	if err != nil {
		logger.Error("Failed to create the score", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return false
	}

	// Grab the information needed for the event
	team, err := r.db.GetTeam(newScore.TeamID)
	if err != nil {
		logger.Error("Failed to get the team record", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return false
	}

	// Send the flag notification
	flag, err := r.db.GetFlag(newScore.FlagID)
	if err != nil {
		logger.Error("Failed to get the flag record", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return false
	}

	r.eventSend("flags", api.EventFlag{Team: *team, Flag: flag, Input: flag.Flag, Value: newScore.Value, Type: "valid"})

	// Send the timeline notification
	total, err := r.db.GetTeamPoints(newScore.TeamID)
	if err != nil {
		logger.Error("Failed to get the team score record", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return false
	}

	score := api.TimelineEntryScore{
		SubmitTime: time.Now(),
		Value:      newScore.Value,
		Total:      total,
	}

	r.eventSend("timeline", api.EventTimeline{TeamID: team.ID, Team: &team.AdminTeamPut.TeamPut, Score: &score, Type: "score-updated"})

	logger.Info("New score entry defined", log15.Ctx{"id": id, "flagid": newScore.FlagID, "teamid": newScore.TeamID, "value": newScore.Value})

	return true
}

func (r *rest) adminGetScore(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	idVar := mux.Vars(request)["id"]

	// Convert the provided id to int
	id, err := strconv.ParseInt(idVar, 10, 64)
	if err != nil {
		logger.Warn("Invalid score ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(400, "Invalid score ID provided", writer, request)
		return
	}

	// Attempt to get the DB record
	score, err := r.db.GetScore(id)
	if err == sql.ErrNoRows {
		logger.Warn("Invalid score ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(404, "Invalid score ID provided", writer, request)
		return
	} else if err != nil {
		logger.Error("Failed to get the score", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	r.jsonResponse(score, writer, request)
}

func (r *rest) adminUpdateScore(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	idVar := mux.Vars(request)["id"]

	// Convert the provided id to int
	id, err := strconv.ParseInt(idVar, 10, 64)
	if err != nil {
		logger.Warn("Invalid score ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(400, "Invalid score ID provided", writer, request)
		return
	}

	// Decode the provided JSON input
	newScore := api.AdminScorePut{}
	err = json.NewDecoder(request.Body).Decode(&newScore)
	if err != nil {
		logger.Warn("Malformed JSON provided", log15.Ctx{"error": err})
		r.errorResponse(400, "Malformed JSON provided", writer, request)
		return
	}

	// Get the current entry
	currentScore, err := r.db.GetScore(id)
	if err == sql.ErrNoRows {
		logger.Warn("Invalid score ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(404, "Invalid score ID provided", writer, request)
		return
	} else if err != nil {
		logger.Error("Failed to get the score entry", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	// Get the initial total
	totalBefore, err := r.db.GetTeamPoints(currentScore.TeamID)
	if err != nil {
		logger.Error("Failed to get the team score record", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	// Get the team
	team, err := r.db.GetTeam(currentScore.TeamID)
	if err != nil {
		logger.Error("Failed to get the team record", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	// Attempt to update the database
	err = r.db.UpdateScore(id, newScore)
	if err == sql.ErrNoRows {
		logger.Warn("Invalid score ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(404, "Invalid score ID provided", writer, request)
		return
	} else if err != nil {
		logger.Error("Failed to update the score", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	// Send the timeline notification
	totalAfter, err := r.db.GetTeamPoints(currentScore.TeamID)
	if err != nil {
		logger.Error("Failed to get the team score record", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	score := api.TimelineEntryScore{
		SubmitTime: time.Now(),
		Value:      totalAfter - totalBefore,
		Total:      totalAfter,
	}

	r.eventSend("timeline", api.EventTimeline{TeamID: team.ID, Team: &team.AdminTeamPut.TeamPut, Score: &score, Type: "score-updated"})

	logger.Info("Score entry updated", log15.Ctx{"id": id, "value": newScore.Value})
}

func (r *rest) adminDeleteScore(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	idVar := mux.Vars(request)["id"]

	// Convert the provided id to int
	id, err := strconv.ParseInt(idVar, 10, 64)
	if err != nil {
		logger.Warn("Invalid score ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(400, "Invalid score ID provided", writer, request)
		return
	}

	// Get the current entry
	currentScore, err := r.db.GetScore(id)
	if err == sql.ErrNoRows {
		logger.Warn("Invalid score ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(404, "Invalid score ID provided", writer, request)
		return
	} else if err != nil {
		logger.Error("Failed to get the score entry", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	// Get the initial total
	totalBefore, err := r.db.GetTeamPoints(currentScore.TeamID)
	if err != nil {
		logger.Error("Failed to get the team score record", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	// Get the team
	team, err := r.db.GetTeam(currentScore.TeamID)
	if err != nil {
		logger.Error("Failed to get the team record", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	// Attempt to delete the DB record
	err = r.db.DeleteScore(id)
	if err == sql.ErrNoRows {
		logger.Warn("Invalid score ID provided", log15.Ctx{"id": idVar})
		r.errorResponse(404, "Invalid score ID provided", writer, request)
		return
	} else if err != nil {
		logger.Error("Failed to delete the score", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	// Send the timeline notification
	totalAfter, err := r.db.GetTeamPoints(currentScore.TeamID)
	if err != nil {
		logger.Error("Failed to get the team score record", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	score := api.TimelineEntryScore{
		SubmitTime: time.Now(),
		Value:      totalAfter - totalBefore,
		Total:      totalAfter,
	}

	r.eventSend("timeline", api.EventTimeline{TeamID: team.ID, Team: &team.AdminTeamPut.TeamPut, Score: &score, Type: "score-updated"})

	logger.Info("Score entry deleted", log15.Ctx{"id": id})
}

func (r *rest) adminClearScores(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	emptyVar := request.FormValue("empty")

	// Confirm the user is sure about it
	if emptyVar != "1" {
		logger.Warn("Scores clear requested without empty=1")
		r.errorResponse(400, "Scores clear requested without empty=1", writer, request)
		return
	}

	// Clear the database entries
	err := r.db.ClearScores()
	if err != nil {
		logger.Error("Failed to clear all scores", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	logger.Info("All scores deleted")
}
