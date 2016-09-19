package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"bitbucket.org/nstratos/anisync/anisync"
)

const imgPlaceholder = "/static/assets/img/placeholder_100x145.png"

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
		return NewAppError(err, "Test sync: could not decode body.", http.StatusBadRequest)
	}

	c := newAnisyncClient(app.httpClient, "", r)

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
		return NewAppError(err, "Test sync: could not encode failures.", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)

	return nil
}

func test1() ([]anisync.Anime, []anisync.Anime) {

	now := time.Now()
	before := now.AddDate(0, 0, -1)
	anime1 := "Death parade"

	anime2 := "Ore monogatari"

	anime3 := "Shingeki no Kyojin"

	anime4 := "Kuroko no basuke"

	malist := []anisync.Anime{
		{
			ID:              1,
			Title:           anime1,
			Rating:          "4.0",
			Image:           imgPlaceholder,
			LastUpdated:     &now,
			Status:          anisync.StatusOnHold,
			EpisodesWatched: 0,
		},
		{
			ID:              3,
			Title:           anime3,
			Rating:          "3.5",
			Image:           imgPlaceholder,
			LastUpdated:     &before,
			EpisodesWatched: 5,
			Rewatching:      false,
		},
		{
			ID:              4,
			Title:           anime4,
			Rating:          "4.5",
			Image:           imgPlaceholder,
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
			Image:           imgPlaceholder,
			LastUpdated:     &now,
			Status:          anisync.StatusOnHold,
			EpisodesWatched: 0,
		},
		{
			ID:          2,
			Title:       anime2,
			Rating:      "4.0",
			Image:       imgPlaceholder,
			LastUpdated: &now,
			Status:      anisync.StatusCurrentlyWatching,
		},
		{
			ID:              3,
			Title:           anime3,
			Rating:          "2.5",
			Image:           imgPlaceholder,
			LastUpdated:     &now,
			EpisodesWatched: 10,
			Rewatching:      true,
		},
		{
			ID:              4,
			Title:           anime4,
			Rating:          "5.0",
			Image:           imgPlaceholder,
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
		err := fmt.Errorf("accounts do not match or unknown test")
		return NewAppError(err, "Test check: Could not run test case.", http.StatusUnauthorized)
	}

	diff := anisync.Compare(malist, hblist)

	// Including MyAnimeList account username in mock response.
	resp := struct {
		MalUsername string
		*anisync.Diff
	}{
		malu,
		diff,
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		return NewAppError(err, "Test check: Could not encode list difference.", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
	return nil
}
