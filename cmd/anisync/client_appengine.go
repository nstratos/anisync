// +build appengine

package main

import (
	"net/http"

	"github.com/nstratos/go-hummingbird/hb"
	"github.com/nstratos/go-myanimelist/mal"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"

	"bitbucket.org/nstratos/anisync/anisync"
)

func newAnisyncClient(httpClient *http.Client, malAgent string, r *http.Request) *anisync.Client {
	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)
	resources := anisync.NewResources(mal.NewClient(client), malAgent, hb.NewClientHTTP(client))
	return anisync.NewClient(resources)
}
