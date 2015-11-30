package anisync_test

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"bitbucket.org/nstratos/anisync/anisync"

	"github.com/nstratos/go-hummingbird/hb"
	"github.com/nstratos/go-myanimelist/mal"
)

const defaultUserAgent = `
	Mozilla/5.0 (X11; Linux x86_64) 
	AppleWebKit/537.36 (KHTML, like Gecko) 
	Chrome/42.0.2311.90 Safari/537.36`

var (
	client *anisync.Client
)

func init() {
	resources := struct {
		*HBClientStub
		*MALClientStub
	}{
		NewHBClientStub(hb.NewClient(nil)),
		NewMALClientStub(mal.NewClient(), defaultUserAgent),
	}
	client = anisync.NewClient(resources)
}

type ResourcesStub struct {
	HBClientStub
	MALClientStub
}

type HBClientStub struct {
	client *hb.Client
}

func NewHBClientStub(hbClient *hb.Client) *HBClientStub {
	return &HBClientStub{client: hb.NewClient(nil)}
}

type MALClientStub struct {
	client *mal.Client
}

func NewMALClientStub(malClient *mal.Client, malAgent string) *MALClientStub {
	c := &MALClientStub{client: mal.NewClient()}
	c.client.SetUserAgent(malAgent)
	return c
}

func (c *MALClientStub) Verify(username, password string) error {
	switch {
	case username == "TestUsername" && password == "TestPassword":
		return nil
	default:
		return fmt.Errorf("wrong password")
	}

}
func (c *MALClientStub) MyAnimeList(username string) (*mal.AnimeList, *mal.Response, error) {
	var (
		animeList *mal.AnimeList
		resp      *mal.Response = &mal.Response{Body: []byte{}, Response: &http.Response{}}
		err       error
	)
	switch {
	case username == "TestUser":
		animeList = &mal.AnimeList{
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
		err = nil
	case username == "TestUserInvalidTime":
		animeList = &mal.AnimeList{
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
		err = nil
	case username == "TestUserInvalidStatus":
		animeList = &mal.AnimeList{
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
		err = nil
	default:
		animeList = &mal.AnimeList{Error: "Invalid username"}
		err = fmt.Errorf("Invalid username")
	}
	return animeList, resp, err
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

func TestClient_GetMyAnimeList_invalidTime(t *testing.T) {
	got, _, err := client.GetMyAnimeList("TestUserInvalidTime")
	if err != nil {
		t.Errorf("GetMyAnimeList with invalid time, instead of skipping, returned error %v", err)
	}
	lastUpdated := time.Date(2015, time.August, 24, 17, 15, 06, 0, time.UTC)
	// Anime from MAL with invalid time are skipped.
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
	// Anime from MAL with invalid time are skipped.
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

func TestClient_VerifyMALCredentials(t *testing.T) {
	err := client.VerifyMALCredentials("TestUsername", "TestPassword")
	if err != nil {
		t.Errorf("VerifyMALCredentials with correct username and password expected to return nil")
	}
}

func TestClient_VerifyMALCredentials_wrongPassword(t *testing.T) {
	err := client.VerifyMALCredentials("TestUser", "WrongTestPassword")
	if err == nil {
		t.Errorf("VerifyMALCredentials with wrong password expected to return err")
	}
}
