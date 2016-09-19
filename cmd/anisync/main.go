// +build !appengine

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var port = flag.String("port", "8080", "server port")

func main() {
	flag.Parse()

	fmt.Println("Starting server at :" + *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatal("ListenandServe:", err)
	}

}
