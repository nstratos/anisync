// +build !appengine

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	var (
		httpAddr = flag.String("http", ":8080", "the host and port on which the server should serve HTTP requests")
	)
	flag.Parse()

	fmt.Println("Starting server at", *httpAddr)
	if err := http.ListenAndServe(*httpAddr, nil); err != nil {
		log.Fatal("ListenandServe:", err)
	}

}
