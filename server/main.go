package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"bitbucket.org/nstratos/anisync/anisync"
	"bitbucket.org/nstratos/anisync/onepic"
)

//go:generate go run generate/includeagent.go

const assetsFolder = "ui/"

var port = flag.String("port", "8080", "server port")

func main() {
	flag.Parse()

	// Preparing ui
	uiHandler := http.FileServer(http.Dir(assetsFolder))
	http.Handle("/static/", http.StripPrefix("/static", uiHandler))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		uiHandler.ServeHTTP(w, r)
	})

	// API handlers
	http.Handle("/api/check", appHandler((check)))
	http.Handle("/api/sync", appHandler((sync)))
	http.Handle("/api/test/check", appHandler((testCheck)))
	http.Handle("/api/getoneimg", appHandler((getOneImg)))

	fmt.Println("Starting server at :" + *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatal("ListenandServe:", err)
	}

}

func check(w http.ResponseWriter, r *http.Request) error {
	hbUsername := r.FormValue("hbUsername")
	malUsername := r.FormValue("malUsername")
	fmt.Println("hbUsername:", hbUsername)
	fmt.Println("malUsername:", malUsername)

	// malAgent is produced by go generate.
	c := anisync.NewClient(malAgent)

	malist, resp, err := c.Anime.ListMAL(malUsername)
	if err != nil {
		if resp.StatusCode == 404 {
			return &appErr{err, fmt.Sprintf("could not get MyAnimeList for user %v", malUsername), http.StatusNotFound}
		} else {
			return &appErr{err, "could not get MyAnimeList", http.StatusInternalServerError}
		}
	}

	hblist, resp, err := c.Anime.ListHB(hbUsername)
	if err != nil {
		if resp.StatusCode == 404 {
			return &appErr{err, fmt.Sprintf("could not get Hummingbird list for user %v", hbUsername), http.StatusNotFound}
		} else {
			return &appErr{err, "could not get Hummingbird list", http.StatusInternalServerError}
		}
	}

	diff := anisync.Compare(malist, hblist)

	bytes, err := json.Marshal(diff)
	if err != nil {
		return &appErr{err, "could not marshal diff", http.StatusInternalServerError}
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
	return nil
}

func getOneImgURL(q string) string {
	return fmt.Sprintf("api/getoneimg?q=%v", url.QueryEscape(q))
}

func getOneImg(w http.ResponseWriter, r *http.Request) error {
	q := r.FormValue("q")
	url, err := onepic.Search(q)
	if err != nil {
		return err
	}
	fmt.Printf("url: %v\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Length", fmt.Sprint(resp.ContentLength))
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	if _, err = io.Copy(w, resp.Body); err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func testCheck(w http.ResponseWriter, r *http.Request) error {
	now := time.Now()
	malist := []anisync.Anime{
		{
			ID:          1,
			Title:       "Death parade",
			Rating:      "4.0",
			Image:       getOneImgURL("Death parade"),
			LastUpdated: &now,
		},
		{
			ID:          2,
			Title:       "Ore monogatari",
			Rating:      "3.0",
			Image:       getOneImgURL("ore monogatari"),
			LastUpdated: &now,
		},
	}
	hblist := []anisync.Anime{
		{
			ID:          1,
			Title:       "Death parade",
			Rating:      "5.0",
			Image:       getOneImgURL("Death parade"),
			LastUpdated: &now,
		},
		{
			ID:          2,
			Title:       "Ore monogatari",
			Rating:      "4.0",
			Image:       getOneImgURL("ore monogatari"),
			LastUpdated: &now,
		},
	}

	diff := anisync.Compare(malist, hblist)

	bytes, err := json.Marshal(diff)
	if err != nil {
		return &appErr{err, "could not marshal diff", http.StatusInternalServerError}
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
	return nil
}
func sync(w http.ResponseWriter, r *http.Request) error {
	return &appErr{nil, "wip", http.StatusNotImplemented}
}

type appErr struct {
	err     error
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e *appErr) Error() string { return fmt.Sprintf("%d %v: %v", e.Code, e.Message, e.err.Error()) }

type appHandler func(http.ResponseWriter, *http.Request) error

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	if err := fn(w, r); err != nil {
		if e, ok := err.(*appErr); ok {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(e.Code)
			if err := json.NewEncoder(w).Encode(e); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	log.Printf("%s\t%s\t%s", r.Method, r.RequestURI, time.Since(start))
}
