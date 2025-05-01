package rest

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/inconshreveable/log15"
	"github.com/lxc/incus/v6/shared/ws"

	"github.com/nsec/askgod/api"
	"github.com/nsec/askgod/internal/utils"
)

var (
	eventHostname  string
	eventsLock     sync.Mutex
	eventListeners = make(map[string]*eventListener)
)

type eventListener struct {
	connection   *websocket.Conn
	messageTypes []string

	active  chan bool
	id      string
	msgLock sync.Mutex
	peer    bool
	teamid  int64
}

func (r *rest) injectEvents(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	// Access control
	if !r.isPeer(request) {
		logger.Warn("Unauthorized attempt to send events")
		r.errorResponse(403, "Forbidden", writer, request)

		return
	}

	// Setup websocket
	conn, err := ws.Upgrader.Upgrade(writer, request, nil)
	if err != nil {
		logger.Error("Failed to setup websocket", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)

		return
	}

	// Process messages
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var rawEvent any
		err = json.Unmarshal(data, &rawEvent)
		if err != nil {
			logger.Error("Received a broken event from peer", log15.Ctx{"error": err})

			continue
		}

		// Handle config reloads
		var apiEvent api.Event
		err = json.Unmarshal(data, &apiEvent)
		if err == nil && apiEvent.Type == "internal" {
			conf, err := r.db.GetConfig()
			if err != nil {
				logger.Error("Failed to get new configuration", log15.Ctx{"error": err})

				continue
			}

			// Save old config
			oldConfig := r.config.ConfigPut
			newConfig := conf

			r.config.ConfigPut = *newConfig
			logger.Info("Config updated", log15.Ctx{"old": oldConfig, "new": newConfig})

			err = r.configHiddenTeams()
			if err != nil {
				logger.Error("Failed to update hidden teams", log15.Ctx{"error": err})

				continue
			}

			continue
		}

		err = r.eventSendRaw(rawEvent)
		if err != nil {
			logger.Error("Failed to relay event from peer", log15.Ctx{"error": err})

			continue
		}
	}
}

func (r *rest) getEvents(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	listener := eventListener{}

	// Get the provided event type
	typeStr := request.FormValue("type")
	if typeStr == "" {
		logger.Warn("Events request without a type")
		r.errorResponse(400, "Missing event type", writer, request)

		return
	}

	if typeStr == "cluster" {
		r.injectEvents(writer, request, logger)

		return
	}

	// Valid the provided types
	eventTypes := strings.Split(typeStr, ",")
	for _, entry := range eventTypes {
		// Make sure that all types are valid
		if !utils.StringInSlice(entry, []string{"timeline", "logging", "flags"}) {
			logger.Warn("Invalid event type", log15.Ctx{"type": entry})
			r.errorResponse(400, "Invalid event type", writer, request)

			return
		}

		// Admin access control
		if utils.StringInSlice(entry, []string{"logging", "flags"}) && !r.hasAccess("admin", request) {
			logger.Warn("Unauthorized attempt to get events", log15.Ctx{"type": entry})
			r.errorResponse(403, "Forbidden", writer, request)

			return
		}
	}

	// Setup websocket
	c, err := ws.Upgrader.Upgrade(writer, request, nil)
	if err != nil {
		logger.Error("Failed to setup websocket", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)

		return
	}

	// Extract the client IP
	ip, err := r.getIP(request)
	if err != nil {
		logger.Error("Failed to get the client's IP", log15.Ctx{"error": err})
		r.errorResponse(500, "Internal Server Error", writer, request)

		return
	}

	// Get the team id
	teamid := int64(0)

	if r.hasAccess("admin", request) {
		teamid = -1
	} else {
		team, err := r.db.GetTeamForIP(*ip)
		if err == nil {
			teamid = team.ID
		}
	}

	// Prepare the listener
	listener.active = make(chan bool, 1)
	listener.connection = c
	listener.id = uuid.New().String()
	listener.messageTypes = eventTypes
	listener.teamid = teamid

	// Add it to the set
	eventsLock.Lock()
	eventListeners[listener.id] = &listener
	eventsLock.Unlock()

	r.logger.Debug("New events listener", log15.Ctx{"uuid": listener.id})

	<-listener.active

	eventsLock.Lock()
	delete(eventListeners, listener.id)
	eventsLock.Unlock()

	_ = listener.connection.Close()
	r.logger.Debug("Disconnected events listener", log15.Ctx{"uuid": listener.id})
}

func (r *rest) eventSend(eventType string, eventMessage any) error {
	event := map[string]any{}
	event["type"] = eventType
	event["timestamp"] = time.Now()
	event["metadata"] = eventMessage

	if eventHostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		eventHostname = hostname
	}
	event["server"] = eventHostname

	return r.eventSendRaw(event)
}

func (r *rest) eventSendRaw(raw any) error {
	body, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	event := api.Event{}
	err = json.Unmarshal(body, &event)
	if err != nil {
		return err
	}

	eventsLock.Lock()
	listeners := eventListeners
	for _, listener := range listeners {
		// Don't re-transmit cluster events
		if event.Server != eventHostname && listener.peer {
			continue
		}

		// Only send the right event types
		if listener.messageTypes != nil && !utils.StringInSlice(event.Type, listener.messageTypes) {
			continue
		}

		// If a team message and hide_others is in effect, restrict broadcast
		if event.Type == "timeline" {
			timeline := api.EventTimeline{}
			err = json.Unmarshal(event.Metadata, &timeline)
			if err != nil {
				return err
			}

			if timeline.TeamID > 0 && listener.teamid != -1 && timeline.TeamID != listener.teamid {
				if r.config.Scoring.HideOthers {
					continue
				}

				if utils.Int64InSlice(timeline.TeamID, r.hiddenTeams) {
					continue
				}
			}
		}

		go func(listener *eventListener, body []byte) {
			if listener == nil {
				return
			}

			listener.msgLock.Lock()
			err = listener.connection.WriteMessage(websocket.TextMessage, body)
			listener.msgLock.Unlock()

			if err != nil {
				listener.active <- false
			}
		}(listener, body)
	}
	eventsLock.Unlock()

	return nil
}

func logContextMap(ctx []any) map[string]string {
	var key string
	ctxMap := map[string]string{}

	for _, entry := range ctx {
		if key == "" {
			var ok bool

			key, ok = entry.(string) //nolint:staticcheck
			if !ok {
				continue
			}
		} else {
			ctxMap[key] = fmt.Sprintf("%v", entry)
			key = ""
		}
	}

	return ctxMap
}

func (r *rest) forwardEvents(peer string) {
	var peerURL string
	if strings.HasPrefix(peer, "https://") {
		peerURL = fmt.Sprintf("wss://%s/1.0/events?type=cluster", strings.TrimPrefix(peer, "https://"))
	} else {
		peerURL = fmt.Sprintf("ws://%s/1.0/events?type=cluster", strings.TrimPrefix(peer, "http://"))
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}

	if r.config.Daemon.HTTPSCertificate != "" {
		cert := r.config.Daemon.HTTPSCertificate
		if !strings.Contains(cert, "\n") && utils.PathExists(cert) {
			content, err := os.ReadFile(cert) //nolint:gosec
			if err != nil {
				r.logger.Error("Failed to read cluster certificate", log15.Ctx{"error": err, "peer": peer})

				return
			}

			cert = string(content)
		}

		caCertPool := tlsConfig.RootCAs
		if caCertPool == nil {
			caCertPool = x509.NewCertPool()
		}

		content := cert
		for content != "" {
			block, remainder := pem.Decode([]byte(content))
			if block == nil {
				r.logger.Error("Failed to decode cluster certificate", log15.Ctx{"peer": peer})

				return
			}

			crt, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				r.logger.Error("Failed to parse cluster certificate", log15.Ctx{"error": err, "peer": peer})

				return
			}

			if !crt.IsCA {
				// Override the ServerName
				if crt.DNSNames != nil {
					tlsConfig.ServerName = crt.DNSNames[0]
				} else {
					tlsConfig.ServerName = crt.Subject.CommonName
				}
			}

			caCertPool.AddCert(crt)

			content = string(remainder)
		}

		// Setup the pool
		tlsConfig.RootCAs = caCertPool
	}

	dialer := websocket.Dialer{
		TLSClientConfig: tlsConfig,
	}

	for i := 0; i < 20; i++ {
		r.logger.Debug("Connecting to cluster peer", log15.Ctx{"peer": peer})

		conn, _, err := dialer.Dial(peerURL, nil) //nolint:bodyclose
		if err != nil {
			r.logger.Warn("Failed to connect to cluster peer", log15.Ctx{"error": err, "peer": peer})
		} else {
			listener := eventListener{
				connection: conn,
				active:     make(chan bool, 1),
				id:         uuid.New().String(),
				peer:       true,
				teamid:     -1,
			}

			eventsLock.Lock()
			eventListeners[listener.id] = &listener
			eventsLock.Unlock()
			r.logger.Info("Connected to cluster peer", log15.Ctx{"peer": peer})

			i = 0
			<-listener.active

			r.logger.Warn("Lost connection with cluster peer", log15.Ctx{"peer": peer})
		}

		time.Sleep(3 * time.Second)
	}
	r.logger.Error("Giving up on cluster peer", log15.Ctx{"peer": peer})
}

// EventsLogHandler represents a log15 handler for the /1.0/events API.
type EventsLogHandler struct{}

// Log send a log message through websocket.
func (EventsLogHandler) Log(rec *log15.Record) error {
	r := rest{}
	_ = r.eventSend("logging", api.EventLogging{
		Message: rec.Msg,
		Level:   rec.Lvl.String(),
		Context: logContextMap(rec.Ctx),
	})

	return nil
}
