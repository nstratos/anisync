// +build debug

package kitsu

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

func dumpRequest(req *http.Request, body bool) {
	dump, err := httputil.DumpRequest(req, body)
	if err != nil {
		log.Println("Request dump failed:", err)
		return
	}
	fmt.Println("------------ Request dump -------------")
	fmt.Println(string(dump))
	fmt.Println("")
}

func dumpResponse(resp *http.Response, body bool) {
	dump, err := httputil.DumpResponse(resp, body)
	if err != nil {
		log.Println("Response dump failed:", err)
		return
	}
	fmt.Println("------------ Response dump ------------")
	fmt.Println(string(dump))
	fmt.Println("")
	fmt.Println("---------------------------------------")
}
