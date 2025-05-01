package rest

import (
	"net/http"

	"github.com/inconshreveable/log15"

	"github.com/nsec/askgod/internal/config"
	"github.com/nsec/askgod/internal/database"
)

type rest struct {
	config      *config.Config
	db          *database.DB
	logger      log15.Logger
	router      *http.ServeMux
	hiddenTeams []int64
}
