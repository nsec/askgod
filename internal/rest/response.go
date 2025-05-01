package rest

import (
	"encoding/json"
	"net/http"

	"github.com/inconshreveable/log15"

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

		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	}
}

func (r *rest) jsonResponse(data any, writer http.ResponseWriter, _ *http.Request) {
	// Set the content type to JSON
	writer.Header().Set("Content-Type", "application/json")

	// Writer the response
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	err := encoder.Encode(data)
	if err != nil {
		r.logger.Error("Failed to marshal response to JSON", log15.Ctx{"error": err})
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)

		return
	}
}

func (*rest) errorResponse(code int, message string, writer http.ResponseWriter, _ *http.Request) {
	// Writer the response
	http.Error(writer, message, code)
}
