package rest

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var metricRequests = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "askgod_requests_total",
	},
)

var metricSubmitTeam = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "askgod_scores_total",
	},
	[]string{"team_id", "type"},
)
