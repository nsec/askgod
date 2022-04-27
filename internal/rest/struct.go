package rest

import (
	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"

	"github.com/nsec/askgod/internal/config"
	"github.com/nsec/askgod/internal/database"
)

type rest struct {
	config      *config.Config
	db          *database.DB
	logger      log15.Logger
	router      *mux.Router
	hiddenTeams []int64
}
