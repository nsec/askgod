package rest

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var metricRequests = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "askgod_requests_total",
		Help: "Total number of requests handled by Askgod",
	},
)

var metricSubmitTeam = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "askgod_scores_total",
		Help: "Scores per team and type",
	},
	[]string{"team_id", "type"},
)
