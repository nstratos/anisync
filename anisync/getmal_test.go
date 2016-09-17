package anisync_test

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/nstratos/go-myanimelist/mal"

	"bitbucket.org/nstratos/anisync/anisync"
)

func (c *MALClientStub) MyAnimeList(username string) (*mal.AnimeList, *mal.Response, error) {
	switch {
	case username == "TestUser":
		animeList := &mal.AnimeList{
			MyInfo: mal.AnimeMyInfo{Name: username},
			Anime: []mal.Anime{
				mal.Anime{
					SeriesAnimeDBID:   1,
					SeriesTitle:       "series title",
					MyWatchedEpisodes: 5,
					MyStatus:          3,            // on-hold
					MyScore:           7,            // Will become 3.5 as Rating.
					MyLastUpdated:     "1440436506", // 2015-08-24 17:15:06 +0000 UTC
					MyRewatching:      "1",
					MyRewatchingEp:    2,
					SeriesImage:       "http://cdn.myanimelist.net/images/anime/1/test-image.jpg",
				},
			},
		}
		resp := &mal.Response{Body: []byte{}, Response: &http.Response{}}
		return animeList, resp, nil
	case username == "TestUserInvalidTime":
		animeList := &mal.AnimeList{
			MyInfo: mal.AnimeMyInfo{Name: username},
			Anime: []mal.Anime{
				mal.Anime{
					SeriesAnimeDBID: 1,
					SeriesTitle:     "title with invalid time",
					MyStatus:        6,
					MyScore:         7,
					MyLastUpdated:   "", // invalid time
				},
				mal.Anime{
					SeriesAnimeDBID: 2,
					SeriesTitle:     "normal title",
					MyStatus:        4,
					MyScore:         9,
					MyLastUpdated:   "1440436506", // 2015-08-24 17:15:06 +0000 UTC
				},
			},
		}
		resp := &mal.Response{Body: []byte{}, Response: &http.Response{}}
		return animeList, resp, nil
	case username == "TestUserInvalidStatus":
		animeList := &mal.AnimeList{
			MyInfo: mal.AnimeMyInfo{Name: username},
			Anime: []mal.Anime{
				mal.Anime{
					SeriesAnimeDBID: 1,
					SeriesTitle:     "title with invalid status",
					MyScore:         8,
					MyLastUpdated:   "1440436506", // 2015-08-24 17:15:06 +0000 UTC
				},
				mal.Anime{
					SeriesAnimeDBID: 2,
					SeriesTitle:     "normal status title",
					MyStatus:        4,
					MyScore:         8,
					MyLastUpdated:   "1440436506", // 2015-08-24 17:15:06 +0000 UTC
				},
			},
		}
		resp := &mal.Response{Body: []byte{}, Response: &http.Response{}}
		return animeList, resp, nil
	case username == "TestNoResponse":
		return nil, nil, fmt.Errorf("no response from test myanimelist server")
	default:
		animeList := &mal.AnimeList{Error: "Invalid username"}
		resp := &mal.Response{Body: []byte{}, Response: &http.Response{}}
		err := fmt.Errorf("Invalid username")
		return animeList, resp, err
	}
}

func TestClient_GetMyAnimeList(t *testing.T) {
	got, _, err := client.GetMyAnimeList("TestUser")
	if err != nil {
		t.Errorf("GetMyAnimeList returned error %v", err)
	}
	lastUpdated := time.Date(2015, time.August, 24, 17, 15, 06, 0, time.UTC)
	want := []anisync.Anime{
		{
			ID:              1,
			Title:           "series title",
			EpisodesWatched: 5,
			Status:          anisync.StatusOnHold,
			Rating:          "3.5",
			LastUpdated:     &lastUpdated,
			Rewatching:      true,
			TimesRewatched:  2,
			Image:           "http://cdn.myanimelist.net/images/anime/1/test-image.jpg",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetMyAnimeList returned %+v, want %+v", got, want)
	}
}

func TestClient_GetMyAnimeList_invalidUsername(t *testing.T) {
	_, _, err := client.GetMyAnimeList("InvalidTestUser")
	if err == nil {
		t.Errorf("GetMyAnimeList for invalid user expected to return err")
	}
}

func TestClient_GetMyAnimeList_noResponse(t *testing.T) {
	_, resp, err := client.GetMyAnimeList("TestNoResponse")
	if err == nil {
		t.Error("GetMyAnimeList for no response expected to return err")
	}
	if resp != nil {
		t.Error("GetMyAnimeList for no response resp = %q, want %q", resp, nil)
	}
}

func TestClient_GetMyAnimeList_invalidTime(t *testing.T) {
	got, _, err := client.GetMyAnimeList("TestUserInvalidTime")
	if err != nil {
		t.Errorf("GetMyAnimeList with invalid time, instead of skipping, returned error %v", err)
	}
	lastUpdated := time.Date(2015, time.August, 24, 17, 15, 06, 0, time.UTC)
	// We are skipping Anime from MyAnimeList.net with invalid time.
	want := []anisync.Anime{
		{
			ID:          2,
			Title:       "normal title",
			Status:      anisync.StatusDropped,
			Rating:      "4.5",
			LastUpdated: &lastUpdated,
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetMyAnimeList returned %+v, want %+v", got, want)
	}

}

func TestClient_GetMyAnimeList_invalidStatus(t *testing.T) {
	got, _, err := client.GetMyAnimeList("TestUserInvalidStatus")
	if err != nil {
		t.Errorf("GetMyAnimeList with invalid status, instead of skipping, returned error %v", err)
	}
	lastUpdated := time.Date(2015, time.August, 24, 17, 15, 06, 0, time.UTC)
	// We are skipping Anime from MyAnimeList.net with invalid time.
	want := []anisync.Anime{
		{
			ID:          2,
			Title:       "normal status title",
			Status:      anisync.StatusDropped,
			Rating:      "4.0",
			LastUpdated: &lastUpdated,
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetMyAnimeList returned %+v, want %+v", got, want)
	}

}
