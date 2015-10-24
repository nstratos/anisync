// TODO:
// * Sync functionality.
// * Disable sync button, enable after check.
// * Search filter for results.
// * Filter results based on tag.
// * Option button to include comments.
// * Integration tests.
// * Full unit tests.
// * Deploy on app engine.
// * Cookies to keep usernames and options.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
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

func testCheck(w http.ResponseWriter, r *http.Request) error {
	now := time.Now()
	before := now.AddDate(0, 0, -1)
	anime1 := "Death parade"
	anime1Pic, err := onepic.Search(anime1)
	if err != nil {
		return err
	}
	anime2 := "Ore monogatari"
	anime2Pic, err := onepic.Search(anime2)
	if err != nil {
		return err
	}
	anime3 := "Shingeki no Kyojin"
	anime3Pic, err := onepic.Search(anime3)
	if err != nil {
		return err
	}
	anime4 := "Kuroko no basuke"
	anime4Pic, err := onepic.Search(anime4)
	if err != nil {
		return err
	}
	malist := []anisync.Anime{
		{
			ID:              1,
			Title:           anime1,
			Rating:          "4.0",
			Image:           anime1Pic,
			LastUpdated:     &now,
			Status:          anisync.StatusOnHold,
			EpisodesWatched: 0,
		},
		{
			ID:              3,
			Title:           anime3,
			Rating:          "3.5",
			Image:           anime3Pic,
			LastUpdated:     &before,
			EpisodesWatched: 5,
			Rewatching:      false,
		},
		{
			ID:              4,
			Title:           anime4,
			Rating:          "4.5",
			Image:           anime4Pic,
			LastUpdated:     &before,
			EpisodesWatched: 6,
			Rewatching:      false,
		},
	}
	hblist := []anisync.Anime{
		{
			ID:              1,
			Title:           anime1,
			Rating:          "4.0",
			Image:           anime1Pic,
			LastUpdated:     &now,
			Status:          anisync.StatusOnHold,
			EpisodesWatched: 0,
		},
		{
			ID:          2,
			Title:       anime2,
			Rating:      "4.0",
			Image:       anime2Pic,
			LastUpdated: &now,
			Status:      anisync.StatusCurrentlyWatching,
		},
		{
			ID:              3,
			Title:           anime3,
			Rating:          "2.5",
			Image:           anime3Pic,
			LastUpdated:     &now,
			EpisodesWatched: 10,
			Rewatching:      true,
		},
		{
			ID:              4,
			Title:           anime4,
			Rating:          "5.0",
			Image:           anime4Pic,
			LastUpdated:     &now,
			EpisodesWatched: 6,
			Rewatching:      true,
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
