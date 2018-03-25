package anisync

import (
	"fmt"
	"strconv"
	"time"

	"github.com/nstratos/go-kitsu/kitsu"
)

const kitsuTimeLayout = "2006-01-02T15:04:05.000Z"

// KitsuClient is a Hummingbird client that contains implementations for all the
// operations that we need from the Hummingbird.me API.
type KitsuClient struct {
	client *kitsu.Client
}

// NewKitsuClient creates a new Hummingbird client.
func NewKitsuClient(client *kitsu.Client) *KitsuClient {
	return &KitsuClient{client: client}
}

func (c *KitsuClient) KitsuAnimeList(userID string) ([]*kitsu.LibraryEntry, *kitsu.Response, error) {
	return c.client.Library.List(
		kitsu.Include("anime"),
		kitsu.Include("anime.mappings"),
		kitsu.Filter("userId", userID),
	)
}

func (c *Client) GetKitsuAnimeList(username string) ([]*Anime, *kitsu.Response, error) {
	entries, resp, err := c.resources.KitsuAnimeList(username)
	if err != nil {
		return nil, resp, err
	}
	var anime []*Anime
	for _, e := range entries {
		a, err := fromKitsuEntry(e)
		if err != nil {
			return nil, resp, err
		}
		anime = append(anime, a)
	}
	return anime, resp, nil
}

func fromKitsuEntry(e *kitsu.LibraryEntry) (*Anime, error) {
	a := &Anime{
		EpisodesWatched: e.Progress,
		Status:          fromKitsuStatus(e.Status),
		Notes:           e.Notes,
		TimesRewatched:  e.ReconsumeCount,
		Rewatching:      e.Reconsuming,
		Rating:          e.Rating,
	}
	//2016-11-12T03:35:00.064Z
	updatedAt, err := time.Parse(kitsuTimeLayout, e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("parsing as %s: %v", kitsuTimeLayout, err)
	}
	a.LastUpdated = &updatedAt
	if e.Anime != nil {
		a.Title = e.Anime.CanonicalTitle
		imgURL, ok := e.Anime.PosterImage["tiny"]
		if ok {
			s, ok := imgURL.(string)
			if ok {
				a.Image = s
			}
		}
		if e.Anime.Mappings != nil {
			for _, m := range e.Anime.Mappings {
				if m.ExternalSite == kitsu.ExternalSiteMALAnime {
					id, err := strconv.Atoi(m.ExternalID)
					if err != nil {
						return nil, fmt.Errorf("converting anime ID: %v", err)
					}
					a.ID = id
				}
			}
		}
	}
	// rating
	//if e.Rating != nil {
	//	if e.Rating.Type == "advanced" {
	//		a.Rating = e.Rating.Value
	//	}
	//}
	return a, nil
}
