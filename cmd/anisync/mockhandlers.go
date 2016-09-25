package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"bitbucket.org/nstratos/anisync/anisync"
)

const imgPlaceholder = "/static/assets/img/placeholder_100x145.png"

func (app *App) handleTestSync(w http.ResponseWriter, r *http.Request) error {
	// Receiving json from POST body.
	t := struct {
		HBUsername  string `json:"hbUsername"`
		MALUsername string `json:"malUsername"`
		MALPassword string `json:"malPassword"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		return NewAppError(err, "Test sync: could not decode body.", http.StatusBadRequest)
	}
	hbu := t.HBUsername
	malu := t.MALUsername

	var malist, hblist []anisync.Anime
	switch {
	case hbu == "test1" && hbu == malu:
		malist, hblist = test1()
	default:
		err := fmt.Errorf("accounts do not match or unknown test")
		return NewAppError(err, "Test sync: Could not run test case.", http.StatusUnauthorized)
	}

	diff := anisync.Compare(malist, hblist)
	syncResult := syncMALAnimeTest(diff)

	// Including MyAnimeList account username in response.
	resp := struct {
		MalUsername string
		Sync        *anisync.SyncResult
		*anisync.Diff
	}{
		t.MALUsername,
		syncResult,
		diff,
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		return NewAppError(err, "Test sync: could not encode response.", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)

	return nil
}

func syncMALAnimeTest(diff *anisync.Diff) *anisync.SyncResult {
	var adds []anisync.AddSuccess
	var addf []anisync.AddFail
	missing := make([]anisync.Anime, len(diff.Missing))
	copy(missing, diff.Missing)
	for i, a := range diff.Missing {
		// even succeed, odd fail
		if i%2 != 0 {
			err := fmt.Errorf("add failure but not really")
			addf = append(addf, anisync.MakeAddFail(a, err))
		}
		adds = append(adds, anisync.AddSuccess{Anime: a})
		diff.UpToDate = append(diff.UpToDate, a)
		// delete
		missing = append(missing[:i], missing[i+1:]...)
	}
	diff.Missing = missing

	var upds []anisync.UpdateSuccess
	var updf []anisync.UpdateFail
	needUpdate := make([]anisync.AniDiff, len(diff.NeedUpdate))
	copy(needUpdate, diff.NeedUpdate)
	for i, d := range diff.NeedUpdate {
		// even succeed, odd fail
		if i%2 != 0 {
			err := fmt.Errorf("update failure but not really")
			updf = append(updf, anisync.MakeUpdateFail(d, err))
			continue
		}
		upds = append(upds, anisync.UpdateSuccess{AniDiff: d})
		diff.UpToDate = append(diff.UpToDate, d.Anime)
		// delete
		needUpdate = append(needUpdate[:i], needUpdate[i+1:]...)
	}
	diff.NeedUpdate = needUpdate

	return &anisync.SyncResult{Adds: adds, AddFails: addf, Updates: upds, UpdateFails: updf}
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
