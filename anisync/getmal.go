package anisync

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/nstratos/go-myanimelist/mal"
)

func (c *Client) GetMyAnimeList(username string) ([]Anime, *http.Response, error) {
	list, resp, err := c.resources.MyAnimeList(username)
	if err != nil {
		return nil, resp.Response, err
	}
	// Silently ignoring bad MAL entries if any.
	anime, _ := fromMALEntries(*list)
	return anime, resp.Response, nil
}

type badMALEntry struct {
	MALAnime mal.Anime
	Error    error
}

func fromMALEntries(malist mal.AnimeList) ([]Anime, []badMALEntry) {
	var anime []Anime
	var fails []badMALEntry
	for _, mala := range malist.Anime {
		a, err := fromMALEntry(mala)
		if err != nil {
			fails = append(fails, badMALEntry{MALAnime: mala, Error: err})
			continue
		}
		anime = append(anime, a)
	}
	return anime, fails
}

func fromMALEntry(mala mal.Anime) (Anime, error) {
	a := Anime{
		ID:              mala.SeriesAnimeDBID,
		Title:           mala.SeriesTitle,
		EpisodesWatched: mala.MyWatchedEpisodes,
		TimesRewatched:  mala.MyRewatchingEp,
		Image:           mala.SeriesImage,
		//Notes:           mala.Comments, // MAL API does not send the comments.
	}
	// Status
	status, err := fromMALStatus(mala.MyStatus)
	if err != nil {
		return Anime{}, fmt.Errorf("no status in Anime(ID: %v, Title: %q) : %v", a.ID, a.Title, err)
	}
	a.Status = status

	// LastUpdated
	lastUpdated, err := fromMALMyLastUpdated(mala.MyLastUpdated)
	if err != nil {
		errfmt := "could not parse mal time of Anime(ID: %v, Title: %q, LastUpdated: %q) : %v"
		parseErr := fmt.Errorf(errfmt, a.ID, a.Title, mala.MyLastUpdated, err)
		return Anime{}, parseErr
	}
	a.LastUpdated = lastUpdated
	// Rating
	score := float64(mala.MyScore) / 2
	a.Rating = fmt.Sprintf("%.1f", score)
	// Rewatching
	if mala.MyRewatching == "1" {
		a.Rewatching = true
	}
	return a, nil
}

func fromMALMyLastUpdated(updated string) (*time.Time, error) {
	i, err := strconv.ParseInt(updated, 10, 64)
	if err != nil {
		return nil, err
	}
	t := time.Unix(i, 0).UTC()
	return &t, nil
}
