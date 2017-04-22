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
MIIHGDCCBQCgAwIBAgIJAOXlbhWTdScOMA0GCSqGSIb3DQEBCwUAMIG4MQswCQYD
VQQGEwJDQTEPMA0GA1UECBMGUXVlYmVjMREwDwYDVQQHEwhNb250cmVhbDERMA8G
A1UEChMITm9ydGhTZWMxIDAeBgNVBAsTF0ludGVybmFsIEluZnJhc3RydWN0dXJl
MScwJQYDVQQDEx5Ob3J0aFNlYyAyMDE1IEludGVybmFsIFJvb3QgQ0ExJzAlBgkq
hkiG9w0BCQEWGG5zZWMtaW5mcmFAbGlzdHMubnNlYy5pbzAeFw0xNTAzMTQxNTI3
NDVaFw0yMDAzMTIxNTI3NDVaMIG4MQswCQYDVQQGEwJDQTEPMA0GA1UECBMGUXVl
YmVjMREwDwYDVQQHEwhNb250cmVhbDERMA8GA1UEChMITm9ydGhTZWMxIDAeBgNV
BAsTF0ludGVybmFsIEluZnJhc3RydWN0dXJlMScwJQYDVQQDEx5Ob3J0aFNlYyAy
MDE1IEludGVybmFsIFJvb3QgQ0ExJzAlBgkqhkiG9w0BCQEWGG5zZWMtaW5mcmFA
bGlzdHMubnNlYy5pbzCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBAObR
ZIfpoopGrAegiegtVeDUO24S8WOLgUHdXuRD71Djt8Hou9JnRBbwZCfPi5q9ywUk
I7bFWG3pqKNWxQAWsjTTgZuYuhV5rYaaNnE5+/aQNYaYKiam5QCWBZ6KPs4DAQGQ
RYVJKdW2Ze/OGy7gBv2yFlBR+B5c5SU734pZzJR1Px82LtsN1RtUeFc2kNi+Z6R8
5Bpe6g57qxWUanuk9Y1j+aacTu/jMWkVGWVzY6p7X44MlT+jbJhbixR55ZNvkM0X
v/7SIdaBJ/0sWJEJv9EJq7nN4M8GfpihvCRPjDL0jnimWg+rqKO1Mc1CScOw7Nfd
bpAUjT6WIxG/yx3DFEoMG8h+kI3KwDvlt0Lz7c+1y9+BQ/4YS2smYPL3VvMroUnB
iW9Pe7+iC+agby6zaVpaF7Tj3RgsUia3T1aYKJT+leMU5wUpTrDbv/IrVfxEcEJf
+i6aaESAW0pX1d4ZpEVKiBDzW0lA0wvtD1aeCdQOeO297BeXQin8zJPQ0xXJkiZv
3xNkufZaP6gU9ska7eQyE+ZzCEVCjzB/0RRc2wskxFFzQlRZFZ+mssGWmhTIzVfB
S5+A5DFtJfmW+2iaP4W/dbmJksigduub7jE6NjAKDEqCbxKfrRwatgGnGlPJZ6C/
vZLBZbGad5fcQsjbfy7thoqvWptpWbDn1VR0ZaejAgMBAAGjggEhMIIBHTAdBgNV
HQ4EFgQUtY7VqEC4yQVucu9EoJJ9atajjdAwge0GA1UdIwSB5TCB4oAUtY7VqEC4
yQVucu9EoJJ9atajjdChgb6kgbswgbgxCzAJBgNVBAYTAkNBMQ8wDQYDVQQIEwZR
dWViZWMxETAPBgNVBAcTCE1vbnRyZWFsMREwDwYDVQQKEwhOb3J0aFNlYzEgMB4G
A1UECxMXSW50ZXJuYWwgSW5mcmFzdHJ1Y3R1cmUxJzAlBgNVBAMTHk5vcnRoU2Vj
IDIwMTUgSW50ZXJuYWwgUm9vdCBDQTEnMCUGCSqGSIb3DQEJARYYbnNlYy1pbmZy
YUBsaXN0cy5uc2VjLmlvggkA5eVuFZN1Jw4wDAYDVR0TBAUwAwEB/zANBgkqhkiG
9w0BAQsFAAOCAgEAW9KSU4cCWXBBu+eVTNBAcEudasqz4UgyHC+mB6cXKGG9pKIt
NRgyBxgXD+M0XcsUoKgad8xhWOwzpFEw2CVd5ARJi6vUeVuMtFbLAMQqRqMma6tF
Q+vKwufACuaWDO69ozKB/WHzwXbIh0KzcnAR1GLx+H7hkr4CXTPcb88rSinaxr9K
GrszKL1iy6T89kkYmdZsrXkboDJ+WPmXh/be2Yx/bC1WZc4fgVuMyRKBEir0ODsZ
xC79LVyUlw5kokzIILRpAhqdN5MGvgLWnhueBTdI4SqKybY6RaGklOSN14fLBNvJ
e174Jq5Pgz/Q51gZz+PyoOE6ZaKKUIhkfmLuGegM8i0O/7CaJKR0R//uDHp7T2lz
pVd/pbPflI256VXWzh0Qdzbp+0Lq9Ec+dVCZ0ey9q4Ql4oySp8BnR6nkn5cf9Tyx
o9LFp/xgMcQLBJCp7DeQphJZQcxCl3EgdXHMVAXvA6X+APRDI5jvL2AEevwQiLEG
5cGSx1y563ppCUbchknHPmDekc03AfYeRx99nlUcnyB/gtIVean+W1S7y+2D7jqB
bETlq0bAr1dyznRlkGGqiGaCX5cFzQoCPTPMy9MaHIezjHxy/NF6likg6YQ99ZMx
+kKJMrIyuigO2pKE0FD2ssDuZbRb61AYU8btSt8c/YykaW1R0Xm2L206TRU=
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
