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
MIICEzCCAZigAwIBAgIUPi0EdQHeIsk8d6YRtUKFMhiVBfswCgYIKoZIzj0EAwMw
NjELMAkGA1UEBhMCQ0ExETAPBgNVBAoTCE5vcnRoU2VjMRQwEgYDVQQDEwtOb3J0
aFNlYyBFMTAeFw0yMjA0MjYyMDU5MDFaFw0zMjA0MjMyMDU5MzFaMDoxCzAJBgNV
BAYTAkNBMREwDwYDVQQKEwhOb3J0aFNlYzEYMBYGA1UEAxMPTm9ydGhTZWMgV2Vi
IEUxMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAETkYX+NPp/ywJ+Oe3nG19uIX8FYn8
vcWQuSMjKeBvZZA5E+1HZ37OyrGHa0rSkLYFUBySY1DfhIr2LzsY/5yExFXZ6aeo
9C6PAEOuPuDYkJXYJdiZWA4fWwrqdP3TF3Z3o2MwYTAOBgNVHQ8BAf8EBAMCAQYw
DwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUgw0LZPdf6v3I6+hSFm402aT/oygw
HwYDVR0jBBgwFoAUw1KpwZaajSJBo1sn7tZ4VO3FEJgwCgYIKoZIzj0EAwMDaQAw
ZgIxAJfZPSeIDEIdeiAanDIARe9HR54oFrP3K3PyR5H7sX7nXb+W94thsj3NiL4p
xwH4GwIxAPMw3XZOYn5OUhhURS0EImK1by78oZKsUsmAPxVP53uhdeuGEQik70ss
9w0DbEF3bw==
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIICDjCCAZSgAwIBAgIULLm//UTiveLdBkjEaeyOndeR9TswCgYIKoZIzj0EAwMw
NjELMAkGA1UEBhMCQ0ExETAPBgNVBAoTCE5vcnRoU2VjMRQwEgYDVQQDEwtOb3J0
aFNlYyBFMTAeFw0yMjA0MjYyMDQxMzRaFw0zMjA0MjMyMDQyMDRaMDYxCzAJBgNV
BAYTAkNBMREwDwYDVQQKEwhOb3J0aFNlYzEUMBIGA1UEAxMLTm9ydGhTZWMgRTEw
djAQBgcqhkjOPQIBBgUrgQQAIgNiAASV/eOLEpPoGd3DOut2vODfbvlEO37k06Ra
yRvkpxecqfc2NhwJAsiz6BExGf/wSIUydlaInBsuKyoFKfUTAcFzZA+YfuT9SVKH
s0O9GwmSWbQjaoR57WUzRh6c+yzRThyjYzBhMA4GA1UdDwEB/wQEAwIBBjAPBgNV
HRMBAf8EBTADAQH/MB0GA1UdDgQWBBTDUqnBlpqNIkGjWyfu1nhU7cUQmDAfBgNV
HSMEGDAWgBTDUqnBlpqNIkGjWyfu1nhU7cUQmDAKBggqhkjOPQQDAwNoADBlAjEA
5ilgqNjfrHYA1ahEgS/yX2QMMV9Eff3tZ61JrqD69HpHCQ4ecxp9iT8Jx/LLvp1d
AjAHiDrGrrFZBM3QoPsnsf1OyvewoqY9euXHxTsgWFj+PgY5ld+EyZ6kG2AJj27s
FfM=
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
			MinVersion: tls.VersionTLS13,
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
