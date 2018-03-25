// +build appengine

package main

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

func httpClientFromRequest(r *http.Request) *http.Client {
	ctx := appengine.NewContext(r)
	return urlfetch.Client(ctx)
}
