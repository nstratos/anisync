package anisync

import (
	"net/http"

	"github.com/nstratos/go-hummingbird/hb"
)

func (c *Client) GetHBAnimeList(username string) ([]Anime, *http.Response, error) {
	entries, resp, err := c.resources.HBAnimeList(username)
	if err != nil {
		return nil, resp, err
	}
	anime := fromHBEntries(entries)
	return anime, resp, nil
}

func fromHBEntries(list []hb.LibraryEntry) []Anime {
	var anime []Anime
	for _, hbe := range list {
		a := fromHBEntry(hbe)
		anime = append(anime, a)
	}
	return anime
}

func fromHBEntry(hbe hb.LibraryEntry) Anime {
	a := Anime{
		EpisodesWatched: hbe.EpisodesWatched,
		Status:          fromHBStatus(hbe.Status),
		LastUpdated:     hbe.UpdatedAt,
		Notes:           hbe.Notes,
		TimesRewatched:  hbe.RewatchedTimes,
		Rewatching:      hbe.Rewatching,
	}
	if hbe.Anime != nil {
		a.ID = hbe.Anime.MALID
		a.Title = hbe.Anime.Title
		a.Image = hbe.Anime.CoverImage
	}
	// rating
	if hbe.Rating != nil {
		if hbe.Rating.Type == "advanced" {
			a.Rating = hbe.Rating.Value
		}
	}
	return a
}
