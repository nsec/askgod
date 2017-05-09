package rest

import (
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"

	"github.com/nsec/askgod/internal/config"
	"github.com/nsec/askgod/internal/database"
)

// AttachFunctions attaches all the REST API functions to the provided router
func AttachFunctions(config *config.Config, router *mux.Router, db *database.DB, logger log15.Logger) error {
	r := rest{
		config: config,
		db:     db,
		logger: logger,
		router: router,
	}

	// Guest API
	r.registerEndpoint("/", "guest", r.getRoot, nil, nil, nil)
	r.registerEndpoint("/1.0", "guest", r.getStatus, nil, nil, nil)

	r.registerEndpoint("/1.0/events", "guest", r.getEvents, r.injectEvents, nil, nil)

	r.registerEndpoint("/1.0/scoreboard", "guest", r.getScoreboard, nil, nil, nil)
	r.registerEndpoint("/1.0/timeline", "guest", r.getTimeline, nil, nil, nil)

	// Team API
	r.registerEndpoint("/1.0/team", "team", r.getTeam, nil, r.updateTeam, nil)
	r.registerEndpoint("/1.0/team/flags", "team", r.getTeamFlags, r.submitTeamFlag, nil, nil)
	r.registerEndpoint("/1.0/team/flags/{id}", "team", r.getTeamFlag, nil, r.updateTeamFlag, nil)

	// Admin API
	r.registerEndpoint("/1.0/config", "admin", r.getConfig, nil, nil, nil)

	r.registerEndpoint("/1.0/flags", "admin", r.adminGetFlags, r.adminCreateFlag, nil, r.adminClearFlags)
	r.registerEndpoint("/1.0/flags/{id}", "admin", r.adminGetFlag, nil, r.adminUpdateFlag, r.adminDeleteFlag)

	r.registerEndpoint("/1.0/scores", "admin", r.adminGetScores, r.adminCreateScore, nil, r.adminClearScores)
	r.registerEndpoint("/1.0/scores/{id}", "admin", r.adminGetScore, nil, r.adminUpdateScore, r.adminDeleteScore)

	r.registerEndpoint("/1.0/teams", "admin", r.adminGetTeams, r.adminCreateTeam, nil, r.adminClearTeams)
	r.registerEndpoint("/1.0/teams/{id}", "admin", r.adminGetTeam, nil, r.adminUpdateTeam, r.adminDeleteTeam)

	// Setup forwarder
	for _, peer := range config.Daemon.ClusterPeers {
		go r.forwardEvents(peer)
	}

	// Listen for config changes
	err := config.RegisterHandler(r.configChanged)
	if err != nil {
		return err
	}

	return nil
}

func (r *rest) registerEndpoint(url string, access string, funcGet, funcPost, funcPut, funcDelete func(writer http.ResponseWriter, request *http.Request, logger log15.Logger)) {
	r.router.HandleFunc(url, func(writer http.ResponseWriter, request *http.Request) {
		if !r.hasAccess(access, request) {
			r.errorResponse(403, "Forbidden", writer, request)
			return
		}

		r.logger.Debug("Request received", log15.Ctx{"method": request.Method, "url": request.URL, "client": request.RemoteAddr})
		logger := r.logger.New("method", request.Method, "url", request.URL, "client", request.RemoteAddr)

		// Process the Origin header
		r.processOrigin(writer, request)

		// Process OPTIONS
		if request.Method == "OPTIONS" {
			return
		}

		switch request.Method {
		case "GET":
			if funcGet != nil {
				funcGet(writer, request, logger)
				return
			}
		case "POST":
			if funcPost != nil {
				funcPost(writer, request, logger)
				return
			}
		case "PUT":
			if funcPut != nil {
				funcPut(writer, request, logger)
				return
			}
		case "DELETE":
			if funcDelete != nil {
				funcDelete(writer, request, logger)
				return
			}
		}

		r.logger.Info("Bad request (not implemented)", log15.Ctx{"method": request.Method, "url": request.URL, "client": request.RemoteAddr})
		r.errorResponse(501, "Not Implemented", writer, request)
		return
	})
}
