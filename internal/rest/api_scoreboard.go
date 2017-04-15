package rest

import (
	"net/http"

	"gopkg.in/inconshreveable/log15.v2"
)

func (r *rest) getScoreboard(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	scoreboard, err := r.db.GetScoreboard()
	if err != nil {
		logger.Error("Failed to get the scoreboard", log15.Ctx{"error": err})
		r.errorResponse(500, "Internal Server Error", writer, request)
		return
	}

	r.jsonResponse(scoreboard, writer, request)
}
