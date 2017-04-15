package rest

import (
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"

	"github.com/nsec/askgod/internal/config"
	"github.com/nsec/askgod/internal/database"
)

type rest struct {
	config *config.Config
	db     *database.DB
	logger log15.Logger
	router *mux.Router
}
