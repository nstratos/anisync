package anisync

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/nstratos/go-myanimelist/mal"
)

func (c *Client) GetMyAnimeList(username string) ([]Anime, *http.Response, error) {
	list, resp, err := c.resources.MyAnimeList(username)
	if err != nil {
		return nil, resp.Response, err
	}
	anime := fromMALEntries(*list)
	sort.Sort(ByID(anime))
	return anime, resp.Response, nil
}

func fromMALEntries(malist mal.AnimeList) []Anime {
	var anime []Anime
	for _, mala := range malist.Anime {
		a, err := fromMALEntry(mala)
		if err != nil {
			log.Printf("Discarded MAL entry: %v", err)
			continue
		}
		anime = append(anime, a)
	}
	return anime
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
