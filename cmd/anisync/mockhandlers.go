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
	var syncFn syncFunc
	switch {
	case hbu == "test1" && hbu == malu:
		malist, hblist, syncFn = test1()
	case hbu == "test2" && hbu == malu:
		malist, hblist, syncFn = test2()
	case hbu == "test3" && hbu == malu:
		malist, hblist, syncFn = test3()
	case hbu == "test4" && hbu == malu:
		malist, hblist, syncFn = test4()
	default:
		err := fmt.Errorf("accounts do not match or unknown test")
		return NewAppError(err, "Test sync: Could not run test case.", http.StatusUnauthorized)
	}

	diff := anisync.Compare(malist, hblist)
	syncResult := syncMALAnimeTest(diff, syncFn)

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

func syncMALAnimeTest(diff *anisync.Diff, syncFn func(index int, anime anisync.Anime) error) *anisync.SyncResult {
	var (
		adds              []anisync.AddSuccess
		addf              []anisync.AddFail
		removeFromMissing []int
	)
	for i, a := range diff.Missing {
		err := syncFn(i, a)
		if err != nil {
			addf = append(addf, anisync.MakeAddFail(a, err))
			continue
		}
		adds = append(adds, anisync.AddSuccess{Anime: a})
		// After gathering the add successes and failures, we also modify diff
		// to make it appear as the sync happened. This includes adding the
		// successes as "up to date" and removing them from "missing".
		diff.UpToDate = append(diff.UpToDate, a)
		// Gathering the IDs of the anime to remove.
		removeFromMissing = append(removeFromMissing, a.ID)
	}
	// Removing the anime.
	for _, id := range removeFromMissing {
		diff.Missing = deleteAnimeByID(diff.Missing, id)
	}

	var (
		upds                 []anisync.UpdateSuccess
		updf                 []anisync.UpdateFail
		removeFromNeedUpdate []int
	)
	for i, d := range diff.NeedUpdate {
		err := syncFn(i, d.Anime)
		if err != nil {
			updf = append(updf, anisync.MakeUpdateFail(d, err))
			continue
		}
		upds = append(upds, anisync.UpdateSuccess{AniDiff: d})
		// After gathering the update successes and failures, we also modify
		// diff to make it appear as the sync happened. This includes adding
		// the successes as "up to date" and removing them from "need update".
		diff.UpToDate = append(diff.UpToDate, d.Anime)
		// Gathering the IDs of the anime to remove.
		removeFromNeedUpdate = append(removeFromNeedUpdate, d.Anime.ID)
	}
	// Removing the the anime.
	for _, id := range removeFromNeedUpdate {
		diff.NeedUpdate = deleteAniDiffByID(diff.NeedUpdate, id)
	}

	return &anisync.SyncResult{Adds: adds, AddFails: addf, Updates: upds, UpdateFails: updf}
}

func deleteAnimeByID(anime []anisync.Anime, id int) []anisync.Anime {
	for i, a := range anime {
		if a.ID == id {
			anime = append(anime[:i], anime[i+1:]...)
		}
	}
	return anime
}

func deleteAniDiffByID(diff []anisync.AniDiff, id int) []anisync.AniDiff {
	for i, d := range diff {
		if d.Anime.ID == id {
			diff = append(diff[:i], diff[i+1:]...)
		}
	}
	return diff
}

// syncFunc is used to indicate which anime will return an error when
// performing a mock sync.
type syncFunc func(index int, anime anisync.Anime) error

func test1() ([]anisync.Anime, []anisync.Anime, syncFunc) {
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
	// One anime fails to update.
	syncFn := func(index int, anime anisync.Anime) error {
		switch anime.ID {
		case 4:
			return fmt.Errorf("anime failed to be updated (but that's normal!)")
		default:
			return nil
		}
		return nil
	}
	return malist, hblist, syncFn
}

func test2() ([]anisync.Anime, []anisync.Anime, syncFunc) {
	now := time.Now()

	anime1 := "One Piece"

	malist := []anisync.Anime{
		{
			ID:              1,
			Title:           anime1,
			Image:           imgPlaceholder,
			LastUpdated:     &now,
			Status:          anisync.StatusOnHold,
			EpisodesWatched: 2,
		},
	}

	hblist := []anisync.Anime{
		{
			ID:              1,
			Title:           anime1,
			Image:           imgPlaceholder,
			LastUpdated:     &now,
			Status:          anisync.StatusOnHold,
			EpisodesWatched: 2,
		},
	}
	// All anime succeed.
	syncFn := func(int, anisync.Anime) error {
		return nil
	}
	return malist, hblist, syncFn
}

func test3() ([]anisync.Anime, []anisync.Anime, syncFunc) {
	now := time.Now()
	before := now.AddDate(0, 0, -1)

	anime1 := "Berserk"
	anime2 := "Cowboy Bebop"

	malist := []anisync.Anime{
		{
			ID:              1,
			Title:           anime1,
			Image:           imgPlaceholder,
			LastUpdated:     &before,
			Status:          anisync.StatusOnHold,
			EpisodesWatched: 0,
		},
	}

	hblist := []anisync.Anime{
		{
			ID:              1,
			Title:           anime1,
			Rating:          "4.0",
			Image:           imgPlaceholder,
			LastUpdated:     &now,
			Status:          anisync.StatusCurrentlyWatching,
			EpisodesWatched: 2,
		},
		{
			ID:              2,
			Title:           anime2,
			Rating:          "3.0",
			Image:           imgPlaceholder,
			LastUpdated:     &now,
			Status:          anisync.StatusOnHold,
			EpisodesWatched: 0,
		},
	}
	// All anime fail.
	syncFn := func(index int, anime anisync.Anime) error {
		switch anime.ID {
		case 1:
			return fmt.Errorf("anime failed to be updated (but that's normal!)")
		case 2:
			return fmt.Errorf("anime failed to be added (but that's normal!)")
		default:
			return nil
		}
	}
	return malist, hblist, syncFn
}

func test4() ([]anisync.Anime, []anisync.Anime, syncFunc) {
	now := time.Now()
	before := now.AddDate(0, 0, -1)

	anime1 := "Death parade"
	anime2 := "Ore monogatari"
	anime3 := "Shingeki no Kyojin"
	anime4 := "Kuroko no basuke"
	anime5 := "One Piece"
	anime6 := "Berserk"
	anime7 := "Cowboy Bebop"
	anime8 := "Mob Psycho 100"

	malist := []anisync.Anime{
		{
			ID:              1,
			Title:           anime1,
			Image:           imgPlaceholder,
			LastUpdated:     &now,
			Status:          anisync.StatusOnHold,
			EpisodesWatched: 0,
		},
		{
			ID:              2,
			Title:           anime2,
			Rating:          "3.0",
			Image:           imgPlaceholder,
			LastUpdated:     &before,
			Status:          anisync.StatusOnHold,
			EpisodesWatched: 0,
		},
		{
			ID:              3,
			Title:           anime3,
			Rating:          "4.0",
			Image:           imgPlaceholder,
			LastUpdated:     &before,
			Status:          anisync.StatusOnHold,
			EpisodesWatched: 0,
		},
		{
			ID:          7,
			Title:       anime7,
			Image:       imgPlaceholder,
			LastUpdated: &before,
			Status:      anisync.StatusOnHold,
		},
		{
			ID:          8,
			Title:       anime8,
			Image:       imgPlaceholder,
			LastUpdated: &now,
			Status:      anisync.StatusOnHold,
		},
	}

	hblist := []anisync.Anime{
		{
			ID:              1,
			Title:           anime1,
			Rating:          "4.0",
			Image:           imgPlaceholder,
			LastUpdated:     &now,
			Status:          anisync.StatusCurrentlyWatching,
			EpisodesWatched: 2,
		},
		{
			ID:              2,
			Title:           anime2,
			Rating:          "3.0",
			Image:           imgPlaceholder,
			LastUpdated:     &now,
			Status:          anisync.StatusOnHold,
			EpisodesWatched: 4,
		},
		{
			ID:              3,
			Title:           anime3,
			Rating:          "4.5",
			Image:           imgPlaceholder,
			LastUpdated:     &now,
			Status:          anisync.StatusPlanToWatch,
			EpisodesWatched: 8,
		},
		{
			ID:          4,
			Title:       anime4,
			Image:       imgPlaceholder,
			LastUpdated: &now,
			Status:      anisync.StatusPlanToWatch,
		},
		{
			ID:          5,
			Title:       anime5,
			Image:       imgPlaceholder,
			LastUpdated: &now,
			Status:      anisync.StatusPlanToWatch,
		},
		{
			ID:          6,
			Title:       anime6,
			Image:       imgPlaceholder,
			LastUpdated: &now,
			Status:      anisync.StatusPlanToWatch,
		},
		{
			ID:          7,
			Title:       anime7,
			Image:       imgPlaceholder,
			LastUpdated: &now,
			Status:      anisync.StatusOnHold,
		},
		{
			ID:          8,
			Title:       anime8,
			Image:       imgPlaceholder,
			LastUpdated: &now,
			Status:      anisync.StatusOnHold,
		},
	}
	// Everything fails.
	syncFn := func(index int, anime anisync.Anime) error {
		switch anime.ID {
		case 1, 2, 3:
			return fmt.Errorf("anime failed to be updated (but that's normal!)")
		case 4, 5, 6:
			return fmt.Errorf("anime failed to be added (but that's normal!)")
		case 7, 8:
			return fmt.Errorf("this error should not appear because 7 and 8 are in sync")
		default:
			return nil
		}
	}
	return malist, hblist, syncFn
}

func (app *App) handleTestCheck(w http.ResponseWriter, r *http.Request) error {
	hbu := r.FormValue("hbUsername")
	malu := r.FormValue("malUsername")

	var malist, hblist []anisync.Anime
	switch {
	case hbu == "test1" && hbu == malu:
		malist, hblist, _ = test1()
	case hbu == "test2" && hbu == malu:
		malist, hblist, _ = test2()
	case hbu == "test3" && hbu == malu:
		malist, hblist, _ = test3()
	case hbu == "test4" && hbu == malu:
		malist, hblist, _ = test4()
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
