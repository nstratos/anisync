// +build !appengine

package main

import (
	"net/http"

	"github.com/nstratos/go-hummingbird/hb"
	"github.com/nstratos/go-myanimelist/mal"

	"bitbucket.org/nstratos/anisync/anisync"
)

func newAnisyncClient(httpClient *http.Client, malAgent string, r *http.Request) *anisync.Client {
	resources := anisync.NewResources(mal.NewClient(mal.HTTPClient(httpClient)), malAgent, hb.NewClient(httpClient))
	return anisync.NewClient(resources)
}
