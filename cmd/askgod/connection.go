package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

var certPinning = map[string]string{
	"https://askgod.nsec": `
-----BEGIN CERTIFICATE-----
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
-----END CERTIFICATE-----
`,
}

func (c *client) setupClient() error {
	u, err := url.ParseRequestURI(c.server)
	if err != nil {
		return err
	}

	var transport *http.Transport

	if u.Scheme == "http" {
		transport = &http.Transport{
			DisableKeepAlives: true,
		}
	} else if u.Scheme == "https" {
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

		cert, ok := certPinning[c.server]
		if ok {
			certBlock, _ := pem.Decode([]byte(cert))
			if certBlock == nil {
				return fmt.Errorf("Failed to load pinned certificate")
			}

			cert, err := x509.ParseCertificate(certBlock.Bytes)
			if err != nil {
				return fmt.Errorf("Failed to parse pinned certificate: %v", err)
			}

			caCertPool := tlsConfig.RootCAs
			if caCertPool == nil {
				caCertPool = x509.NewCertPool()
			}

			caCertPool.AddCert(cert)
			tlsConfig.RootCAs = caCertPool
		}

		transport = &http.Transport{
			TLSClientConfig:   tlsConfig,
			DisableKeepAlives: true,
		}
	} else {
		return fmt.Errorf("Unsupported server URL: %s", c.server)
	}

	c.http = &http.Client{
		Transport: transport,
	}

	return nil
}

func (c *client) queryStruct(method string, path string, data interface{}, target interface{}) error {
	var req *http.Request
	var err error

	url := fmt.Sprintf("%s/1.0%s", c.server, path)

	// Get a new HTTP request setup
	if data != nil {
		// Encode the provided data
		buf := bytes.Buffer{}
		err := json.NewEncoder(&buf).Encode(data)
		if err != nil {
			return err
		}

		// Some data to be sent along with the request
		req, err = http.NewRequest(method, url, &buf)
		if err != nil {
			return err
		}

		// Set the encoding accordingly
		req.Header.Set("Content-Type", "application/json")
	} else {
		// No data to be sent along with the request
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return err
		}
	}

	// Send the request
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		content, err := ioutil.ReadAll(resp.Body)
		if err == nil && string(content) != "" {
			return fmt.Errorf("%s", strings.TrimSpace(string(content)))
		}

		return fmt.Errorf("%s: %s", url, resp.Status)
	}

	// Decode the response
	if target != nil {
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&target)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *client) websocket(path string) (*websocket.Conn, error) {
	// Generate the URL
	var url string
	if strings.HasPrefix(c.server, "https://") {
		url = fmt.Sprintf("wss://%s/1.0%s", strings.TrimPrefix(c.server, "https://"), path)
	} else {
		url = fmt.Sprintf("ws://%s/1.0%s", strings.TrimPrefix(c.server, "http://"), path)
	}

	// Grab the http transport handler
	httpTransport := c.http.Transport.(*http.Transport)

	// Setup a new websocket dialer based on it
	dialer := websocket.Dialer{
		TLSClientConfig: httpTransport.TLSClientConfig,
		Proxy:           httpTransport.Proxy,
	}

	// Establish the connection
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	return conn, err
}
