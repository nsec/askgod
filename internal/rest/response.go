package rest

import (
	"encoding/json"
	"net/http"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/nsec/askgod/internal/utils"
)

func (r *rest) processOrigin(writer http.ResponseWriter, request *http.Request) {
	origin := request.Header.Get("Origin")
	if origin != "" {
		if utils.StringInSlice(origin, r.config.Daemon.AllowedOrigins) {
			writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else if utils.StringInSlice("*", r.config.Daemon.AllowedOrigins) {
			writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
	}
}

func (r *rest) jsonResponse(data interface{}, writer http.ResponseWriter, request *http.Request) {
	// Set the content type to JSON
	writer.Header().Set("Content-Type", "application/json")

	// Process the Origin header
	r.processOrigin(writer, request)

	// Writer the response
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	err := encoder.Encode(data)
	if err != nil {
		r.logger.Error("Failed to marshal response to JSON", log15.Ctx{"error": err})
		http.Error(writer, "Internal Server Error", 500)
		return
	}

	return
}

func (r *rest) errorResponse(code int, message string, writer http.ResponseWriter, request *http.Request) {
	// Process the Origin header
	r.processOrigin(writer, request)

	// Writer the response
	http.Error(writer, message, code)

	return
}
