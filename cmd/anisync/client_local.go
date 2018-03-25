// +build !appengine

package main

import (
	"net/http"
)

func httpClientFromRequest(r *http.Request) *http.Client {
	return nil
}
