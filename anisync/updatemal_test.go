package anisync_test

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/nstratos/anisync/anisync"

	"github.com/nstratos/go-myanimelist/mal"
)

const (
	validAnimeID = iota + 1 // We consider ID 0 as invalid.
	notFoundAnimeID
)

func (c *MALClientStub) UpdateMALAnimeEntry(id int, entry mal.AnimeEntry) (*mal.Response, error) {
	switch {
	case id == validAnimeID:
		return &mal.Response{Body: []byte{}, Response: &http.Response{}}, nil
	case id == notFoundAnimeID:
		return &mal.Response{Body: []byte{}, Response: &http.Response{}}, fmt.Errorf("anime not found")
	default:
		return &mal.Response{Body: []byte{}, Response: &http.Response{}}, fmt.Errorf("invalid ID")
	}
}

func TestClient_UpdateMALAnime(t *testing.T) {
	anime := anisync.Anime{
		ID:         validAnimeID,
		Status:     anisync.Completed,
		Rewatching: true,
		Rating:     "4.5",
	}
	err := client.UpdateMALAnime(anime)
	if err != nil {
		t.Errorf("UpdateMALAnime returned error %v", err)
	}
}

//func TestClient_UpdateMALAnime_invalidStatus(t *testing.T) {
//	anime := anisync.Anime{ID: validAnimeID}
//	err := client.UpdateMALAnime(anime)
//	if err == nil {
//		t.Errorf("UpdateMALAnime with invalid status expected to return err")
//	}
//}

func TestClient_UpdateMALAnime_invalidID(t *testing.T) {
	anime := anisync.Anime{Status: anisync.OnHold}
	err := client.UpdateMALAnime(anime)
	if err == nil {
		t.Errorf("UpdateMALAnime with invalid ID expected to return err")
	}
}

func (c *MALClientStub) AddMALAnimeEntry(id int, entry mal.AnimeEntry) (*mal.Response, error) {
	switch {
	case id == validAnimeID:
		return &mal.Response{Body: []byte{}, Response: &http.Response{}}, nil
	case id == notFoundAnimeID:
		return &mal.Response{Body: []byte{}, Response: &http.Response{}}, fmt.Errorf("anime not found")
	default:
		return &mal.Response{Body: []byte{}, Response: &http.Response{}}, fmt.Errorf("invalid ID")
	}
}

func TestClient_AddMALAnime(t *testing.T) {
	anime := anisync.Anime{
		ID:         validAnimeID,
		Status:     anisync.Completed,
		Rewatching: true,
		Rating:     "4.5",
	}
	err := client.AddMALAnime(anime)
	if err != nil {
		t.Errorf("AddMALAnime returned error %v", err)
	}
}

//func TestClient_AddMALAnime_invalidStatus(t *testing.T) {
//	anime := anisync.Anime{ID: validAnimeID}
//	err := client.AddMALAnime(anime)
//	if err == nil {
//		t.Errorf("AddMALAnime with invalid status expected to return err")
//	}
//}

func TestClient_AddMALAnime_invalidID(t *testing.T) {
	anime := anisync.Anime{Status: anisync.OnHold}
	err := client.AddMALAnime(anime)
	if err == nil {
		t.Errorf("AddMALAnime with invalid ID expected to return err")
	}
}

var syncTests = []struct {
	name       string
	diff       anisync.Diff
	syncResult *anisync.SyncResult
}{
	{
		"one update success and one update fail",
		anisync.Diff{
			Left: []anisync.Anime{
				{
					ID:     validAnimeID,
					Title:  "Anime1",
					Rating: "3.5",
					Status: anisync.OnHold,
				},
				{
					ID:     notFoundAnimeID,
					Title:  "Anime2",
					Rating: "1.0",
					Status: anisync.Dropped,
				},
			},
			Right: []anisync.Anime{
				{
					ID:     validAnimeID,
					Title:  "Anime1",
					Rating: "4.5",
					Status: anisync.OnHold,
				},
				{
					ID:     notFoundAnimeID,
					Title:  "Anime2",
					Rating: "2.0",
					Status: anisync.Dropped,
				},
			},
			NeedUpdate: []anisync.AniDiff{
				{
					Anime: anisync.Anime{
						ID:     validAnimeID,
						Title:  "Anime1",
						Rating: "4.5",
						Status: anisync.OnHold,
					},
					Rating: &anisync.RatingDiff{Got: "3.5", Want: "4.5"},
				},
				{
					Anime: anisync.Anime{
						ID:     notFoundAnimeID,
						Title:  "Anime2",
						Rating: "2.0",
						Status: anisync.Dropped,
					},
					Rating: &anisync.RatingDiff{Got: "1.0", Want: "2.0"},
				},
			},
		},
		&anisync.SyncResult{
			Updates: []anisync.UpdateSuccess{
				{
					AniDiff: anisync.AniDiff{
						Anime: anisync.Anime{
							ID:     validAnimeID,
							Title:  "Anime1",
							Rating: "4.5",
							Status: anisync.OnHold,
						},
						Rating: &anisync.RatingDiff{Got: "3.5", Want: "4.5"},
					},
				},
			},
			UpdateFails: []anisync.UpdateFail{
				{
					AniDiff: anisync.AniDiff{
						Anime: anisync.Anime{
							ID:     notFoundAnimeID,
							Title:  "Anime2",
							Rating: "2.0",
							Status: anisync.Dropped,
						},
						Rating: &anisync.RatingDiff{Got: "1.0", Want: "2.0"},
					},
					Error:  fmt.Errorf("anime not found"),
					Reason: "anime not found",
				},
			},
		},
	},
	{
		"one add success and one add fail",
		anisync.Diff{
			Left: []anisync.Anime{},
			Right: []anisync.Anime{
				{
					ID:     validAnimeID,
					Title:  "Anime1",
					Rating: "4.5",
					Status: anisync.OnHold,
				},
				{
					ID:     notFoundAnimeID,
					Title:  "Anime2",
					Rating: "2.0",
					Status: anisync.Dropped,
				},
			},
			Missing: []anisync.Anime{
				{
					ID:     validAnimeID,
					Title:  "Anime1",
					Rating: "4.5",
					Status: anisync.OnHold,
				},
				{
					ID:     notFoundAnimeID,
					Title:  "Anime2",
					Rating: "2.0",
					Status: anisync.Dropped,
				},
			},
		},
		&anisync.SyncResult{
			Adds: []anisync.AddSuccess{
				{
					Anime: anisync.Anime{
						ID:     validAnimeID,
						Title:  "Anime1",
						Rating: "4.5",
						Status: anisync.OnHold,
					},
				},
			},
			AddFails: []anisync.AddFail{
				{
					Anime: anisync.Anime{
						ID:     notFoundAnimeID,
						Title:  "Anime2",
						Rating: "2.0",
						Status: anisync.Dropped,
					},
					Error:  fmt.Errorf("anime not found"),
					Reason: "anime not found",
				},
			},
		},
	},
}

func TestClient_SyncMALAnime(t *testing.T) {
	for i, tt := range syncTests {
		got := client.SyncMALAnime(tt.diff)
		if want := tt.syncResult; !reflect.DeepEqual(got, want) {
			t.Errorf("SyncMALAnime test #%d %q returned \n%+v, want \n%+v", i, tt.name, got, want)
		}
	}
}
