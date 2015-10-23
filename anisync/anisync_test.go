package anisync

import (
	"reflect"
	"testing"
	"time"
)

func TestCompare_Rating(t *testing.T) {
	left := []Anime{
		{ID: 1, Title: "Anime1", Rating: "3.0"},
		{ID: 2, Title: "Anime2", Rating: "4.0"},
	}
	right := []Anime{
		{ID: 1, Title: "Anime1", Rating: "4.0"},
		{ID: 2, Title: "Anime2", Rating: "5.0"},
	}
	want := &Diff{
		Left:  left,
		Right: right,
		NeedUpdate: []AniDiff{
			{
				Anime:  right[0],
				Rating: &Rating{Got: "3.0", Want: "4.0"},
			},
			{
				Anime:  right[1],
				Rating: &Rating{Got: "4.0", Want: "5.0"},
			},
		},
	}
	got := Compare(left, right)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Compare Rating diff doesn't match, got: \n%+v, want:\n%+v", got, want)
	}

}

func TestCompare_LastUpdate(t *testing.T) {
	now := time.Now()
	before := now.AddDate(0, 0, -1)
	left := []Anime{
		{ID: 1, Title: "Anime1", LastUpdated: &before},
		{ID: 2, Title: "Anime2", LastUpdated: &before},
	}
	right := []Anime{
		{ID: 1, Title: "Anime1", LastUpdated: &now},
		{ID: 2, Title: "Anime2", LastUpdated: &now},
	}
	want := []AniDiff{
		{
			Anime:       right[0],
			LastUpdated: &LastUpdated{Got: before, Want: now},
		},
		{
			Anime:       right[1],
			LastUpdated: &LastUpdated{Got: before, Want: now},
		},
	}
	got := Compare(left, right).NeedUpdate
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Compare LastUpdated diff doesn't match, got: \n%+v, want:\n%+v", got, want)
	}

}
