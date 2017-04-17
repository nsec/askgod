package rest

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lxc/lxd/shared"
	"github.com/pborman/uuid"
	"gopkg.in/inconshreveable/log15.v2"

	"github.com/nsec/askgod/api"
	"github.com/nsec/askgod/internal/utils"
)

var eventHostname string
var eventsLock sync.Mutex
var eventListeners = make(map[string]*eventListener)

type eventListener struct {
	connection   *websocket.Conn
	messageTypes []string

	active  chan bool
	id      string
	msgLock sync.Mutex
	peer    bool
}

func (r *rest) injectEvents(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	// Access control
	if !r.isPeer(request) {
		logger.Warn("Unauthorized attempt to send events")
		r.errorResponse(403, "Forbidden", writer, request)
		return
	}

	// Setup websocket
	conn, err := shared.WebsocketUpgrader.Upgrade(writer, request, nil)
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

		var event interface{}
		err = json.Unmarshal(data, &event)
		if err != nil {
			logger.Error("Received a broken event from peer", log15.Ctx{"error": err})
			continue
		}

		err = eventSendRaw(event)
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
	c, err := shared.WebsocketUpgrader.Upgrade(writer, request, nil)
	if err != nil {
		logger.Error("Failed to setup websocket", log15.Ctx{"error": err})
		r.errorResponse(500, fmt.Sprintf("%v", err), writer, request)
		return
	}

	// Prepare the listener
	listener.active = make(chan bool, 1)
	listener.connection = c
	listener.id = uuid.NewRandom().String()
	listener.messageTypes = eventTypes

	// Add it to the set
	eventsLock.Lock()
	eventListeners[listener.id] = &listener
	eventsLock.Unlock()

	r.logger.Debug("New events listener", log15.Ctx{"uuid": listener.id})

	<-listener.active

	eventsLock.Lock()
	delete(eventListeners, listener.id)
	eventsLock.Unlock()

	listener.connection.Close()
	r.logger.Debug("Disconnected events listener", log15.Ctx{"uuid": listener.id})
}

func eventSend(eventType string, eventMessage interface{}) error {
	event := map[string]interface{}{}
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

	return eventSendRaw(event)
}

func eventSendRaw(event interface{}) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	eventsLock.Lock()
	listeners := eventListeners
	for _, listener := range listeners {
		if event.(map[string]interface{})["server"].(string) != eventHostname && listener.peer {
			continue
		}

		if listener.messageTypes != nil && !utils.StringInSlice(event.(map[string]interface{})["type"].(string), listener.messageTypes) {
			continue
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

func logContextMap(ctx []interface{}) map[string]string {
	var key string
	ctxMap := map[string]string{}

	for _, entry := range ctx {
		if key == "" {
			key = entry.(string)
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
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA},
		PreferServerCipherSuites: true,
	}

	if r.config.Daemon.HTTPSCertificate != "" {
		cert := r.config.Daemon.HTTPSCertificate
		if !strings.Contains(cert, "\n") && utils.PathExists(cert) {
			content, err := ioutil.ReadFile(cert)
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
		tlsConfig.BuildNameToCertificate()
	}

	dialer := websocket.Dialer{
		TLSClientConfig: tlsConfig,
	}

	for i := 0; i < 20; i++ {
		r.logger.Debug("Connecting to cluster peer", log15.Ctx{"peer": peer})

		conn, _, err := dialer.Dial(peerURL, nil)
		if err != nil {
			r.logger.Warn("Failed to connect to cluster peer", log15.Ctx{"error": err, "peer": peer})
		} else {
			listener := eventListener{
				connection: conn,
				active:     make(chan bool, 1),
				id:         uuid.NewRandom().String(),
				peer:       true,
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

// EventsLogHandler represents a log15 handler for the /1.0/events API
type EventsLogHandler struct {
}

// Log send a log message through websocket
func (h EventsLogHandler) Log(r *log15.Record) error {
	eventSend("logging", api.EventLogging{
		Message: r.Msg,
		Level:   r.Lvl.String(),
		Context: logContextMap(r.Ctx)})
	return nil
}
