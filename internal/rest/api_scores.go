package rest

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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
	// Decode the provided JSON input
	newScore := api.AdminScorePost{}
	err := json.NewDecoder(request.Body).Decode(&newScore)
	if err != nil {
		logger.Warn("Malformed JSON provided", log15.Ctx{"error": err})
		r.errorResponse(400, "Malformed JSON provided", writer, request)
		return
	}

	// Attempt to update the database
	id, err := r.db.CreateScore(newScore)
	if err != nil {
		logger.Error("Failed to create the score", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	// Grab the information needed for the event
	team, err := r.db.GetTeam(newScore.TeamID)
	if err != nil {
		logger.Error("Failed to get the team record", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	flag, err := r.db.GetFlag(newScore.FlagID)
	if err != nil {
		logger.Error("Failed to get the flag record", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	eventSend("flags", api.EventFlag{Team: *team, Flag: flag, Input: flag.Flag, Value: newScore.Value, Result: "valid"})

	logger.Info("New score entry defined", log15.Ctx{"id": id, "flagid": newScore.FlagID, "teamid": newScore.TeamID, "value": newScore.Value})
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

	// Attempt to get the DB record
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

	logger.Info("Score entry deleted", log15.Ctx{"id": id})
}
