package rest

import (
	"net"
	"net/http"
	"net/url"
	"strings"

	"gopkg.in/inconshreveable/log15.v2"
)

func (r *rest) getIP(request *http.Request) (*net.IP, error) {
	// Get the client IP
	clientIP, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		r.logger.Error("Unable to parse client address", log15.Ctx{"address": request.RemoteAddr, "error": err})
		return nil, err
	}

	ip := net.ParseIP(clientIP)
	if ip == nil {
		r.logger.Error("Unable to parse client IP", log15.Ctx{"ip": clientIP, "error": err})
		return nil, err
	}

	return &ip, nil
}

func (r *rest) hasAccess(level string, request *http.Request) bool {
	// Check for cluster peers
	if len(r.config.Daemon.ClusterPeers) != 0 && r.isPeer(request) {
		return true
	}

	// Get the IP
	ip, err := r.getIP(request)
	if err != nil {
		return false
	}

	// Check if admin
	for _, entry := range r.config.Subnets.Admins {
		_, subnet, err := net.ParseCIDR(entry)
		if err != nil {
			r.logger.Error("Unable to parse configured subnet", log15.Ctx{"subnet": entry, "error": err})
			continue
		}

		if subnet.Contains(*ip) {
			return true
		}
	}

	if level == "admin" {
		return false
	}

	// Check if team
	for _, entry := range r.config.Subnets.Teams {
		_, subnet, err := net.ParseCIDR(entry)
		if err != nil {
			r.logger.Error("Unable to parse configured subnet", log15.Ctx{"subnet": entry, "error": err})
			continue
		}

		if subnet.Contains(*ip) {
			return true
		}
	}

	if level == "team" {
		return false
	}

	// Check if guest
	for _, entry := range r.config.Subnets.Guests {
		_, subnet, err := net.ParseCIDR(entry)
		if err != nil {
			r.logger.Error("Unable to parse configured subnet", log15.Ctx{"subnet": entry, "error": err})
			continue
		}

		if subnet.Contains(*ip) {
			return true
		}
	}

	r.logger.Warn("Unauthorized access", log15.Ctx{"method": request.Method, "url": request.URL, "client": ip.String()})

	return false
}

func (r *rest) isPeer(request *http.Request) bool {
	ip, err := r.getIP(request)
	if err != nil {
		return false
	}

	names, _ := net.LookupAddr(ip.String())

	for _, peer := range r.config.Daemon.ClusterPeers {
		u, err := url.ParseRequestURI(peer)
		if err != nil {
			r.logger.Error("Unable to parse peer address", log15.Ctx{"peer": peer, "error": err})
			continue
		}

		host, _, err := net.SplitHostPort(u.Host)
		if err != nil {
			r.logger.Error("Unable to parse peer host", log15.Ctx{"peer": peer, "error": err})
			continue
		}

		for _, name := range names {
			if strings.ToLower(strings.TrimSuffix(name, ".")) == strings.ToLower(host) {
				return true
			}
		}

		peerIP := net.ParseIP(strings.Trim(u.Host, "[]"))
		if peerIP != nil {
			if peerIP.Equal(*ip) {
				return true
			}
		}
	}

	return false
}
