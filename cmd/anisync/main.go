package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
}

const assetsFolder = "ui/"

func run() error {
	var (
		httpAddr = flag.String("http", ":"+getenv("PORT", "8080"), "the host and port on which the server should serve HTTP requests")
	)
	flag.Parse()

	// Preparing ui
	uiHandler := http.FileServer(http.Dir(assetsFolder))
	http.Handle("/static/", http.StripPrefix("/static", uiHandler))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		uiHandler.ServeHTTP(w, r)
	})

	app := &App{
		httpClient: http.DefaultClient,
	}

	// API handlers
	http.Handle("/api/check", appHandler(app.handleCheck))
	http.Handle("/api/sync", appHandler(app.handleSync))
	http.Handle("/api/mal-verify", appHandler(app.handleMALVerify))
	http.Handle("/api/mock/check", appHandler(app.handleTestCheck))
	http.Handle("/api/mock/sync", appHandler(app.handleTestSync))
	http.Handle("/api/mock/mal-verify", appHandler(app.handleTestMALVerify))

	log.Println("Starting server at", *httpAddr)
	if err := http.ListenAndServe(*httpAddr, nil); err != nil {
		return fmt.Errorf("ListenAndServe: %v", err)
	}
	return nil
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
