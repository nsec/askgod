package rest

import (
	"net/http"

	"gopkg.in/inconshreveable/log15.v2"
)

func (r *rest) getTimeline(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	timeline, err := r.db.GetTimeline()
	if err != nil {
		logger.Error("Failed to get the timeline", log15.Ctx{"error": err})
		r.errorResponse(500, "Internal Server Error", writer, request)
		return
	}

	r.jsonResponse(timeline, writer, request)
}
