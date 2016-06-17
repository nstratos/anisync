package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"bitbucket.org/nstratos/anisync/anisync"
	"bitbucket.org/nstratos/anisync/onepic"
)

func mockUpdateMAL(diff *anisync.Diff) ([]*anisync.Fail, error) {
	fails0 := []*anisync.Fail{
		{
			Anime: anisync.Anime{
				ID:              1,
				Title:           "Ore monogatari",
				Rating:          "4.0",
				Status:          anisync.StatusOnHold,
				EpisodesWatched: 0,
			},
			Error: fmt.Errorf("something went wrong"),
		},
	}
	fails1 := []*anisync.Fail{
		{
			Anime: anisync.Anime{
				ID:              4,
				Title:           "Kuroko no basuke",
				Rating:          "5.0",
				EpisodesWatched: 6,
				Rewatching:      true,
			}, Error: fmt.Errorf("misdirection overflow"),
		},
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	switch r.Intn(3) {
	case 0:
		return fails0, fmt.Errorf("one failed")
	case 1:
		return fails1, fmt.Errorf("one or more failed")
	case 2:
		return nil, nil
	}

	return nil, nil
}

func (app *App) handleTestSync(w http.ResponseWriter, r *http.Request) error {
	// Receiving json from POST body.
	t := struct {
		HBUsername  string `json:"hbUsername"`
		MALUsername string `json:"malUsername"`
		MALPassword string `json:"malPassword"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		return &appErr{nil, "sync: could not decode body", http.StatusBadRequest}
	}

	c := newAnisyncClient(app.httpClient, app.malAgent, r)

	diff, err := getDiff(c, t.MALUsername, t.HBUsername)
	if err != nil {
		return err
	}

	//err = c.VerifyMALCredentials(t.MALUsername, t.MALPassword)
	//if err != nil {
	//	return &appErr{err, "could not verify MAL credentials", http.StatusUnauthorized}
	//}

	var allFails []*anisync.Fail
	//fails, uerr := c.Anime.UpdateMAL(*diff)
	fails, uerr := mockUpdateMAL(diff)
	allFails = append(allFails, fails...)

	//fails, aerr := c.Anime.AddMAL(*diff)
	fails, aerr := mockUpdateMAL(diff)
	allFails = append(allFails, fails...)

	report := struct {
		Fails   []*anisync.Fail
		Message string
	}{
		allFails,
		"hello",
	}

	if uerr == nil && aerr == nil {
		report.Message = "wow much luck"
	}

	bytes, err := json.Marshal(report)
	if err != nil {
		return &appErr{err, "could not marshal failures", http.StatusInternalServerError}
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)

	return nil
}

func getOnePic(animeName string) string {
	url, err := onepic.Search(animeName)
	if err != nil {
		log.Println("onepic:", err)
		return "/static/assets/img/placeholder_100x145.png"
	}
	return url
}

func test1() ([]anisync.Anime, []anisync.Anime) {

	now := time.Now()
	before := now.AddDate(0, 0, -1)
	anime1 := "Death parade"
	anime1Pic := getOnePic(anime1)

	anime2 := "Ore monogatari"
	anime2Pic := getOnePic(anime2)

	anime3 := "Shingeki no Kyojin"
	anime3Pic := getOnePic(anime3)

	anime4 := "Kuroko no basuke"
	anime4Pic := getOnePic(anime4)

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

	return malist, hblist
}

func (app *App) handleTestCheck(w http.ResponseWriter, r *http.Request) error {
	hbu := r.FormValue("hbUsername")
	malu := r.FormValue("malUsername")

	var malist, hblist []anisync.Anime
	switch {
	case hbu == "test1" && hbu == malu:
		malist, hblist = test1()
	default:
		err := fmt.Errorf("Accounts do not match or unknown test.")
		return &appErr{err, err.Error(), http.StatusUnauthorized}
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
