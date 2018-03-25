package anisync_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/nstratos/anisync/anisync"
)

var (
	now    = time.Now()
	before = now.AddDate(0, 0, -1)
)
var compareTests = []struct {
	name string
	*anisync.Diff
}{
	{name: "NeedUpdate rating", Diff: &anisync.Diff{
		Left:  []anisync.Anime{{ID: 1, Title: "Anime1", Rating: "3.0"}, {ID: 2, Title: "Anime2", Rating: "4.0"}},
		Right: []anisync.Anime{{ID: 1, Title: "Anime1", Rating: "4.0"}, {ID: 2, Title: "Anime2", Rating: "5.0"}},
		NeedUpdate: []anisync.AniDiff{
			{
				Anime:  anisync.Anime{ID: 1, Title: "Anime1", Rating: "4.0"},
				Rating: &anisync.RatingDiff{Got: "3.0", Want: "4.0"},
			},
			{
				Anime:  anisync.Anime{ID: 2, Title: "Anime2", Rating: "5.0"},
				Rating: &anisync.RatingDiff{Got: "4.0", Want: "5.0"},
			},
		},
	}},
	{name: "NeedUpdate last updated (uncertain)", Diff: &anisync.Diff{
		Left:  []anisync.Anime{{ID: 1, Title: "Anime1", LastUpdated: &before}, {ID: 2, Title: "Anime2", LastUpdated: &before}},
		Right: []anisync.Anime{{ID: 1, Title: "Anime1", LastUpdated: &now}, {ID: 2, Title: "Anime2", LastUpdated: &now}},
		Uncertain: []anisync.AniDiff{
			{
				Anime:       anisync.Anime{ID: 1, Title: "Anime1", LastUpdated: &now},
				LastUpdated: &anisync.LastUpdatedDiff{Got: before, Want: now},
			},
			{
				Anime:       anisync.Anime{ID: 2, Title: "Anime2", LastUpdated: &now},
				LastUpdated: &anisync.LastUpdatedDiff{Got: before, Want: now},
			},
		},
	}},
	{name: "NeedUpdate (Not handled case MyAnimeList -> Hummingbird)", Diff: &anisync.Diff{
		Left:     []anisync.Anime{{ID: 1, Title: "Anime1", LastUpdated: &now}},
		Right:    []anisync.Anime{{ID: 1, Title: "Anime1", LastUpdated: &before}},
		UpToDate: []anisync.Anime{{ID: 1, Title: "Anime1", LastUpdated: &before}},
	}},
	{name: "NeedUpdate status", Diff: &anisync.Diff{
		Left:  []anisync.Anime{{ID: 1, Title: "Anime1", Status: anisync.Current}},
		Right: []anisync.Anime{{ID: 1, Title: "Anime1", Status: anisync.Completed}},
		NeedUpdate: []anisync.AniDiff{
			{
				Anime:  anisync.Anime{ID: 1, Title: "Anime1", Status: anisync.Completed},
				Status: &anisync.StatusDiff{Got: anisync.Current, Want: anisync.Completed},
			},
		},
	}},
	{name: "NeedUpdate episodes watched", Diff: &anisync.Diff{
		Left:  []anisync.Anime{{ID: 1, Title: "Anime1", EpisodesWatched: 2}},
		Right: []anisync.Anime{{ID: 1, Title: "Anime1", EpisodesWatched: 5}},
		NeedUpdate: []anisync.AniDiff{
			{
				Anime:           anisync.Anime{ID: 1, Title: "Anime1", EpisodesWatched: 5},
				EpisodesWatched: &anisync.EpisodesWatchedDiff{Got: 2, Want: 5},
			},
		},
	}},
	{name: "NeedUpdate rewatching", Diff: &anisync.Diff{
		Left:  []anisync.Anime{{ID: 1, Title: "Anime1", Rewatching: false}},
		Right: []anisync.Anime{{ID: 1, Title: "Anime1", Rewatching: true}},
		NeedUpdate: []anisync.AniDiff{
			{
				Anime:      anisync.Anime{ID: 1, Title: "Anime1", Rewatching: true},
				Rewatching: &anisync.RewatchingDiff{Got: false, Want: true},
			},
		},
	}},
	{name: "Missing", Diff: &anisync.Diff{
		Left:     []anisync.Anime{{ID: 1, Title: "Anime1"}},
		Right:    []anisync.Anime{{ID: 1, Title: "Anime1"}, {ID: 2, Title: "Anime2"}},
		UpToDate: []anisync.Anime{{ID: 1, Title: "Anime1"}},
		Missing:  []anisync.Anime{{ID: 2, Title: "Anime2"}},
	}},
	{name: "UpToDate", Diff: &anisync.Diff{
		Left:     []anisync.Anime{{ID: 1, Title: "Anime1", LastUpdated: &now}},
		Right:    []anisync.Anime{{ID: 1, Title: "Anime1", LastUpdated: &now}},
		UpToDate: []anisync.Anime{{ID: 1, Title: "Anime1", LastUpdated: &now}},
	}},
	{name: "UpToDate (no rating)", Diff: &anisync.Diff{
		Left:     []anisync.Anime{{ID: 1, Title: "Anime1", Rating: "0.0"}},
		Right:    []anisync.Anime{{ID: 1, Title: "Anime1", Rating: ""}},
		UpToDate: []anisync.Anime{{ID: 1, Title: "Anime1", Rating: ""}},
	}},
	// This case practically can never happen since MAL always sends score 0
	// and compare is meant to be used with left as MAL and right as HB. But
	// just in case that compare it is used with left and right reversed, maybe
	// in the future.
	{name: "UpToDate (no rating reverse)", Diff: &anisync.Diff{
		Left:     []anisync.Anime{{ID: 1, Title: "Anime1", Rating: "0.0"}, {ID: 2, Title: "Anime2", Rating: ""}},
		Right:    []anisync.Anime{{ID: 1, Title: "Anime1", Rating: ""}, {ID: 2, Title: "Anime2", Rating: "0.0"}},
		UpToDate: []anisync.Anime{{ID: 1, Title: "Anime1", Rating: ""}, {ID: 2, Title: "Anime2", Rating: "0.0"}},
	}},
	{name: "wrong order", Diff: &anisync.Diff{
		Left:     []anisync.Anime{{ID: 2, Title: "Anime2"}, {ID: 1, Title: "Anime1"}},
		Right:    []anisync.Anime{{ID: 2, Title: "Anime2"}, {ID: 1, Title: "Anime1"}},
		UpToDate: []anisync.Anime{{ID: 2, Title: "Anime2"}, {ID: 1, Title: "Anime1"}},
	}},
}

func TestCompare(t *testing.T) {
	for i, diff := range compareTests {
		got := anisync.Compare(diff.Left, diff.Right)
		if !reflect.DeepEqual(got, diff.Diff) {
			t.Errorf("Compare test %d:%q did not produce expected diff.", i, diff.name)
		}
	}
}
