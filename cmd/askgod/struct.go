package main

import (
	"net/http"
)

type client struct {
	http       *http.Client
	server     string
	realServer string
}
