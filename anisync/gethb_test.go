package anisync_test

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/nstratos/anisync/anisync"

	"github.com/nstratos/go-hummingbird/hb"
)

func (c *HBClientStub) HBAnimeList(username string) ([]hb.LibraryEntry, *http.Response, error) {
	switch username {
	case "TestUser":
		updatedAt := time.Date(2015, time.December, 01, 01, 27, 01, 0, time.UTC)
		entries := []hb.LibraryEntry{
			{
				Status:          hb.StatusPlanToWatch,
				Rating:          &hb.LibraryEntryRating{Type: "advanced", Value: "4.5"},
				EpisodesWatched: 5,
				UpdatedAt:       &updatedAt,
				Rewatching:      true,
				RewatchedTimes:  2,
				Anime: &hb.Anime{
					Title:      "anime title",
					MALID:      56,
					CoverImage: "https://static.hummingbird.me/anime/poster_images/000/007/622/large/b0012149_5229cf3c7f4ee.jpg",
				},
			},
		}
		resp := &http.Response{}
		return entries, resp, nil
	default:
		entries := []hb.LibraryEntry{}
		resp := &http.Response{}
		err := fmt.Errorf("Invalid username")
		return entries, resp, err
	}
}

func TestClient_GetHBAnimeList(t *testing.T) {
	got, _, err := client.GetHBAnimeList("TestUser")
	if err != nil {
		t.Errorf("GetHBAnimeList returned error %v", err)
	}
	lastUpdated := time.Date(2015, time.December, 01, 01, 27, 01, 0, time.UTC)
	want := []anisync.Anime{
		{
			ID:              56,
			Status:          anisync.Planned,
			Rating:          "4.5",
			EpisodesWatched: 5,
			LastUpdated:     &lastUpdated,
			Rewatching:      true,
			TimesRewatched:  2,
			Title:           "anime title",
			Image:           "https://static.hummingbird.me/anime/poster_images/000/007/622/large/b0012149_5229cf3c7f4ee.jpg",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetHBAnimeList returned \n%+v, want \n%+v", got, want)
	}
}

func TestClient_GetHBAnimeList_invalidUsername(t *testing.T) {
	_, _, err := client.GetHBAnimeList("InvalidTestUser")
	if err == nil {
		t.Errorf("GetHBAnimeList for invalid user expected to return err")
	}
}
