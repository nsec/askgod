package daemon

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/armon/go-proxyproto"
	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/nsec/askgod/internal/config"
	"github.com/nsec/askgod/internal/database"
	"github.com/nsec/askgod/internal/rest"
	"github.com/nsec/askgod/internal/utils"
)

// Daemon is the main struct for all daemon functions.
type Daemon struct {
	// Configuration
	configPath string
	config     *config.Config

	// Database
	db *database.DB

	// Request router
	router *mux.Router

	// Log handler
	logger log15.Logger
}

// NewDaemon returns an initialized Daemon struct.
func NewDaemon(configPath string) (*Daemon, error) {
	d := Daemon{
		configPath: configPath,
		logger:     log15.New(),
	}

	log.SetOutput(io.Discard)

	return &d, nil
}

// Run starts the daemon.
func (d *Daemon) Run() error {
	d.logger.Info("Starting askgod daemon")

	// Read the configuration
	conf, err := config.ReadConfigFile(d.configPath, true, d.logger.New("module", "config"))
	if err != nil {
		return err
	}

	d.config = conf

	// Setup the log handler
	logHandlers := []log15.Handler{}
	logHandlers = append(logHandlers, log15.StreamHandler(os.Stderr, log15.TerminalFormat()))

	logHandlers = append(logHandlers, rest.EventsLogHandler{})

	if d.config.Daemon.LogFile != "" {
		logHandlers = append(logHandlers, log15.Must.FileHandler(d.config.Daemon.LogFile, log15.LogfmtFormat()))
	}

	logLevel := d.config.Daemon.LogLevel
	if logLevel == "" {
		logLevel = "info"
	}

	logFilter, err := log15.LvlFromString(logLevel)
	if err != nil {
		return err
	}

	d.logger.SetHandler(log15.LvlFilterHandler(logFilter, log15.MultiHandler(logHandlers...)))

	// Setup the database
	db, err := database.Connect(
		d.config.Database.Driver,
		d.config.Database.Host,
		d.config.Database.Username,
		d.config.Database.Password,
		d.config.Database.Name,
		d.config.Database.Connections,
		d.config.Database.TLS,
		d.logger.New("module", "database"))
	if err != nil {
		return err
	}

	d.db = db

	// Load the rest of the config
	dbConf, err := d.db.GetConfig()
	if err != nil && errors.Is(err, database.ErrEmptyConfig) {
		d.logger.Info("Config is not found in database. Adding it from the YAML configuration.")
		err := d.db.UpdateConfig(d.config.ConfigPut)
		if err != nil {
			d.logger.Info("Failed to add config.")

			return err
		}

		dbConf, err = d.db.GetConfig()
		if err != nil {
			return err
		}
	} else if err != nil {
		d.logger.Info("Failed to get config.")

		return err
	}

	d.config.ConfigPut = *dbConf

	// Setup the REST API
	d.router = mux.NewRouter()
	err = rest.AttachFunctions(
		d.config,
		d.router,
		d.db,
		d.logger.New("module", "rest"))
	if err != nil {
		return err
	}

	// HTTP
	chServers := make(chan error, 1)
	if d.config.Daemon.HTTPPort > 0 {
		// Prepare the TCP socket
		socket, err := net.Listen("tcp", fmt.Sprintf(":%d", d.config.Daemon.HTTPPort))
		if err != nil {
			return err
		}

		// Wrap for HAProxy
		if d.config.Daemon.HAProxyHeader {
			socket = &proxyproto.Listener{Listener: socket}
		}

		d.logger.Info("Binding HTTP", log15.Ctx{"port": d.config.Daemon.HTTPPort})
		go func() {
			server := &http.Server{
				Handler: d.router,

				ReadTimeout:  10 * time.Second,
				WriteTimeout: 0,
			}

			err := server.Serve(socket)
			if err != nil {
				chServers <- err
				close(chServers)
			}
		}()
	}

	if d.config.Daemon.HTTPSPort > 0 {
		// Load the X509 certificates
		cert := d.config.Daemon.HTTPSCertificate
		if !strings.Contains(cert, "\n") && utils.PathExists(cert) {
			content, err := os.ReadFile(cert) //nolint:gosec
			if err != nil {
				return err
			}

			cert = string(content)
		}

		key := d.config.Daemon.HTTPSKey
		if !strings.Contains(key, "\n") && utils.PathExists(key) {
			content, err := os.ReadFile(key) //nolint:gosec
			if err != nil {
				return err
			}

			key = string(content)
		}

		x509, err := tls.X509KeyPair([]byte(cert), []byte(key))
		if err != nil {
			return err
		}

		// Setup a strict TLS config
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{x509},
			MinVersion:   tls.VersionTLS13,
		}

		// Prepare the TCP socket
		socket, err := net.Listen("tcp", fmt.Sprintf(":%d", d.config.Daemon.HTTPSPort))
		if err != nil {
			return err
		}

		// Wrap for HAProxy
		if d.config.Daemon.HAProxyHeader {
			socket = &proxyproto.Listener{Listener: socket}
		}

		// Wrap for TLS
		socket = tls.NewListener(socket, tlsConfig)

		d.logger.Info("Binding HTTPs", log15.Ctx{"port": d.config.Daemon.HTTPSPort})
		go func() {
			server := &http.Server{
				Handler: d.router,

				ReadTimeout:  10 * time.Second,
				WriteTimeout: 0,
			}

			err := server.Serve(socket)
			if err != nil {
				chServers <- err
				close(chServers)
			}
		}()
	}

	if d.config.Daemon.PrometheusPort > 0 {
		// Prepare the TCP socket
		socket, err := net.Listen("tcp", fmt.Sprintf(":%d", d.config.Daemon.PrometheusPort))
		if err != nil {
			return err
		}

		d.logger.Info("Binding Prometheus", log15.Ctx{"port": d.config.Daemon.PrometheusPort})
		go func() {
			router := mux.NewRouter()
			router.Handle("/metrics", promhttp.Handler())

			server := &http.Server{
				Handler: d.router,

				ReadTimeout:  10 * time.Second,
				WriteTimeout: 0,
			}

			err := server.Serve(socket)
			if err != nil {
				chServers <- err
				close(chServers)
			}
		}()
	}

	err = <-chServers
	if err != nil {
		return err
	}

	return nil
}
