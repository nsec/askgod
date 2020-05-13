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
MIIHdDCCBVygAwIBAgIBATANBgkqhkiG9w0BAQsFADCBtzELMAkGA1UEBhMCQ0Ex
DzANBgNVBAgTBlF1ZWJlYzERMA8GA1UEBxMITW9udHJlYWwxETAPBgNVBAoTCE5v
cnRoU2VjMSAwHgYDVQQLExdJbnRlcm5hbCBJbmZyYXN0cnVjdHVyZTEmMCQGA1UE
AxMdTm9ydGhTZWMgMjAyMCBJbnRlcm5hbCBXZWIgQ0ExJzAlBgkqhkiG9w0BCQEW
GG5zZWMtaW5mcmFAbGlzdHMubnNlYy5pbzAeFw0yMDA1MTMxNzU3NDdaFw0yMjA1
MTMxNzU3NDdaMIGlMQswCQYDVQQGEwJDQTEPMA0GA1UECBMGUXVlYmVjMREwDwYD
VQQHEwhNb250cmVhbDERMA8GA1UEChMITm9ydGhTZWMxIDAeBgNVBAsTF0ludGVy
bmFsIEluZnJhc3RydWN0dXJlMRQwEgYDVQQDEwthc2tnb2QubnNlYzEnMCUGCSqG
SIb3DQEJARYYbnNlYy1pbmZyYUBsaXN0cy5uc2VjLmlvMIICIjANBgkqhkiG9w0B
AQEFAAOCAg8AMIICCgKCAgEAyOQlE0ic4r+bKAWJWyk3ZpeJjyUDR7ZTxBiwNJKK
PTqA2l/0qII/3gsBNh45WAcHS/yq5EayFp1M05qKhMwyYuFus0Fk44KYRd0sbigg
/OqA15ubSu+5t8BJI8S001abDYve531yWg6/5Hm56spKOf6uG/Cq/gqYiFZ8eXwW
VFDZ7UN8uMVe5JNB/HWHrsbrSBgaj/jj116OfWxvScZYtfiO2acjcXDIpjKRO/A3
BypPjNMKRLCSn2x/9VL8jEzIyOvuSSoPk0RVNfxYs8o95PVjZ26bBhhZmNVWv7Gg
9NLQO6W84r3uQIDIDjuJR3LOeW1VVOooDtWfRdMyhfyFCfeht3XAASqrNefoQ8zh
POKw6IGLc0qH603e50v5QbP8Cz6Kw41SmPab37mBFdxtcPMKepM/gtzHsZrrMXzT
UNJaR2IWBbE9tJ+flFLN+uP8++jxK6hVaYszrkd2ILvBCKhwUdB1/ZkN2eUI0nTW
2B+F0YDsOCn3u6ZDrXq00DuQp+qG56ppblHQW6w0Jgd6ALxiq19qlA+DSHy2FBqg
tXTlE6lj5eFr7VEkRxWd7tkse3J6SCUk0b16nCcmeImsK/+n0m6wy8huVz8FtGxe
TIDswOHYUxF8IbSetfCBVWmlm9Ism2cMS5xZapirbi5jZdZqdQ6aMyntj+J5zOl7
31UCAwEAAaOCAZkwggGVMAkGA1UdEwQCMAAwEQYJYIZIAYb4QgEBBAQDAgZAMDQG
CWCGSAGG+EIBDQQnFiVFYXN5LVJTQSBHZW5lcmF0ZWQgU2VydmVyIENlcnRpZmlj
YXRlMB0GA1UdDgQWBBSoeOtQGMpljCjTVxFoEMlYjwqTpzCB5QYDVR0jBIHdMIHa
gBTVJoVZo3FUnxYqS4epqsthA4bgpaGBvqSBuzCBuDELMAkGA1UEBhMCQ0ExDzAN
BgNVBAgTBlF1ZWJlYzERMA8GA1UEBxMITW9udHJlYWwxETAPBgNVBAoTCE5vcnRo
U2VjMSAwHgYDVQQLExdJbnRlcm5hbCBJbmZyYXN0cnVjdHVyZTEnMCUGA1UEAxMe
Tm9ydGhTZWMgMjAyMCBJbnRlcm5hbCBSb290IENBMScwJQYJKoZIhvcNAQkBFhhu
c2VjLWluZnJhQGxpc3RzLm5zZWMuaW+CAQUwEwYDVR0lBAwwCgYIKwYBBQUHAwEw
CwYDVR0PBAQDAgWgMBYGA1UdEQQPMA2CC2Fza2dvZC5uc2VjMA0GCSqGSIb3DQEB
CwUAA4ICAQDSdo2Nr2JQubTNDuZeMDle5PLijwI8f2P6+Bem+Hd36mo/dbre6cfz
0oUMZlCKYQ6P8tI4AkkrgboEFgBo3qgKbMH4s7NTFLtZ6pfvdEo6KHdFrc59Bkz4
IV2QPoCQiWMItvGz3myZuETIn7PXOaZJQ36Xs/3Rr3u7hrJgrlQFsJeANjmKJ9Jj
mRVZBbVdqDs9qy8TnfFijIGN3h6vnD4NL8nSlWdBsFXKyYFTRJcdLK3F12+y9npf
1gNoUBo9jHHm05C3Ch2AGQZZw8qGYmV8h9lbcdBdJd8lah0x6hUwskKxCkGHgWGK
UBeotvOLcaB9MsGd8bVAIoPt2dZBkFZsF/igWGgN65R4OgKUeFNUOtSd0FuJ/OE5
TI/l7iPGswSxfvLGW9Rq6Im/+4stIlwRV4kXy46xzqdDIO1VhBjuDd+OAX2aiSjK
V6gsKaRYeZFMioO5LJQbC0cmfVcngu0zoA0tWI0jAAPR0aFmt5HkZrf28EYb5yIN
iYqNyZB4Luwbe+ntGb7+Q3KWi6lnBxb7pRbc+eGF10QL0ewqXxH3odCSK9izHqX4
BQ32g0Syh9mfyCwrekHc7h71Ha+Nv8YwZKuI9kate8ghydTAejTYwwBBGnGe7VUh
bLRJA5lgaJwxAp8sgn29pMYAY/I3tB2mmTNePVDpDnzBKg0YyQbdgw==
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
