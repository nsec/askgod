package rest

import (
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/inconshreveable/log15"

	"github.com/nsec/askgod/internal/config"
	"github.com/nsec/askgod/internal/database"
)

var clusterPeers []string

// AttachFunctions attaches all the REST API functions to the provided router.
func AttachFunctions(conf *config.Config, router *http.ServeMux, db *database.DB, logger log15.Logger) error {
	r := rest{
		config: conf,
		db:     db,
		logger: logger,
		router: router,
	}

	// Update the list of hidden teams
	err := r.configHiddenTeams()
	if err != nil {
		return err
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
	r.registerEndpoint("/1.0/config", "admin", r.getConfig, nil, r.updateConfig, nil)

	r.registerEndpoint("/1.0/flags", "admin", r.adminGetFlags, r.adminCreateFlag, nil, r.adminClearFlags)
	r.registerEndpoint("/1.0/flags/{id}", "admin", r.adminGetFlag, nil, r.adminUpdateFlag, r.adminDeleteFlag)

	r.registerEndpoint("/1.0/scores", "admin", r.adminGetScores, r.adminCreateScore, nil, r.adminClearScores)
	r.registerEndpoint("/1.0/scores/{id}", "admin", r.adminGetScore, nil, r.adminUpdateScore, r.adminDeleteScore)

	r.registerEndpoint("/1.0/teams", "admin", r.adminGetTeams, r.adminCreateTeam, nil, r.adminClearTeams)
	r.registerEndpoint("/1.0/teams/{id}", "admin", r.adminGetTeam, nil, r.adminUpdateTeam, r.adminDeleteTeam)

	// Setup forwarder
	for _, peer := range conf.Daemon.ClusterPeers {
		u, err := url.ParseRequestURI(peer)
		if err != nil {
			r.logger.Error("Unable to parse peer address", log15.Ctx{"peer": peer, "error": err})

			return err
		}

		host, _, err := net.SplitHostPort(u.Host)
		if err != nil {
			r.logger.Error("Unable to parse peer host", log15.Ctx{"peer": peer, "error": err})

			return err
		}

		peerIP := net.ParseIP(strings.Trim(host, "[]"))
		if peerIP != nil {
			clusterPeers = append(clusterPeers, peerIP.String())
		} else {
			addr, err := net.LookupHost(host)
			if err != nil {
				r.logger.Error("Unable to resolve peer to addr", log15.Ctx{"peer": peer, "error": err})

				return err
			}
			clusterPeers = append(clusterPeers, addr...)
		}

		go r.forwardEvents(peer)
	}

	return nil
}

func (r *rest) registerEndpoint(u string, access string, funcGet, funcPost, funcPut, funcDelete func(writer http.ResponseWriter, request *http.Request, logger log15.Logger)) {
	r.router.HandleFunc(u, func(writer http.ResponseWriter, request *http.Request) {
		metricRequests.Inc()

		if !r.hasAccess(access, request) {
			r.errorResponse(403, "Forbidden", writer, request)

			return
		}

		r.logger.Debug("Request received", log15.Ctx{"method": request.Method, "url": request.URL, "client": request.RemoteAddr})
		logger := r.logger.New("method", request.Method, "url", request.URL, "client", request.RemoteAddr)

		// Process the Origin header
		r.processOrigin(writer, request)

		// Process OPTIONS
		if request.Method == http.MethodOptions {
			return
		}

		switch request.Method {
		case http.MethodGet:
			if funcGet != nil {
				funcGet(writer, request, logger)

				return
			}
		case http.MethodPost:
			if funcPost != nil {
				funcPost(writer, request, logger)

				return
			}
		case http.MethodPut:
			if funcPut != nil {
				funcPut(writer, request, logger)

				return
			}
		case http.MethodDelete:
			if funcDelete != nil {
				funcDelete(writer, request, logger)

				return
			}
		}

		r.logger.Info("Bad request (not implemented)", log15.Ctx{"method": request.Method, "url": request.URL, "client": request.RemoteAddr})
		r.errorResponse(501, "Not Implemented", writer, request)
	})
}
