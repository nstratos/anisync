// +build !debug

package kitsu

import "net/http"

func dumpRequest(req *http.Request, body bool) {}

func dumpResponse(resp *http.Response, body bool) {}
