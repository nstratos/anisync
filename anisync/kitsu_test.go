package anisync

import (
	"fmt"
	"net/http"
	"time"

	"github.com/nstratos/go-kitsu/kitsu"
)

type KitsuClientStub struct {
	client *kitsu.Client
}

func NewKitsuClientStub(kitsuClient *kitsu.Client) *KitsuClientStub {
	return &KitsuClientStub{client: kitsu.NewClient(nil)}
}

func (c *KitsuClientStub) KitsuAnimeList(username string) ([]*kitsu.LibraryEntry, *kitsu.Response, error) {
	switch username {
	case "foo@bar.com":
		updatedAt := time.Date(2015, time.December, 01, 01, 27, 01, 0, time.UTC)
		entries := []*kitsu.LibraryEntry{
			{
				Status:         kitsu.LibraryEntryStatusPlanned,
				Rating:         "4.5",
				Progress:       5,
				UpdatedAt:      updatedAt.Format(kitsuTimeLayout),
				Reconsuming:    true,
				ReconsumeCount: 2,
				Anime: &kitsu.Anime{
					CanonicalTitle: "anime title",
					ID:             "56",
					CoverImage:     map[string]interface{}{"tiny": "https://static.hummingbird.me/anime/poster_images/000/007/622/large/b0012149_5229cf3c7f4ee.jpg"},
				},
			},
		}
		resp := &kitsu.Response{Response: &http.Response{}}
		return entries, resp, nil
	default:
		entries := []*kitsu.LibraryEntry{}
		resp := &kitsu.Response{Response: &http.Response{}}
		err := fmt.Errorf("Invalid username")
		return entries, resp, err
	}
}
