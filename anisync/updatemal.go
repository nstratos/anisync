package anisync

import (
	"fmt"
	"log"
	"math"
	"strconv"

	"github.com/nstratos/go-myanimelist/mal"
)

type Fail struct {
	Anime Anime
	Error error
}

// UpdateMAL gets the difference between the two anime lists and updates the
// the ones that need updating to MyAnimeList based on the values of the
// Hummingbird list.
func (s *AnimeService) UpdateMAL(diff Diff) ([]Fail, error) {
	var failure error
	var updf []Fail
	for _, d := range diff.NeedUpdate {
		err := s.UpdateMALAnime(d.Anime)
		if err != nil {
			failure = fmt.Errorf("failed to update an entry")
			updf = append(updf, Fail{Anime: d.Anime, Error: err})
		}
	}
	return updf, failure
}

// AddMAL gets the difference between the two anime lists and adds the missing
// anime to the MyAnimeList based on the values of the Hummingbird list.
func (s *AnimeService) AddMAL(diff Diff) ([]Fail, error) {
	var failure error
	var addf []Fail
	for _, d := range diff.Missing {
		err := s.AddMALAnime(d)
		if err != nil {
			failure = fmt.Errorf("failed to add an entry")
			addf = append(addf, Fail{Anime: d, Error: err})
		}
	}
	return addf, failure
}

func (s *AnimeService) UpdateMALAnime(a Anime) error {
	e, err := toMALEntry(a)
	if err != nil {
		return err
	}
	printAnimeUpdate(a, "updating")
	_, err = s.client.mal.Anime.Update(a.ID, e)
	if err != nil {
		return err
	}
	return nil
}

func (s *AnimeService) AddMALAnime(a Anime) error {
	e, err := toMALEntry(a)
	if err != nil {
		return err
	}
	printAnimeUpdate(a, "adding")
	_, err = s.client.mal.Anime.Add(a.ID, e)
	if err != nil {
		return err
	}
	return nil
}

func toMALEntry(a Anime) (mal.AnimeEntry, error) {
	e := mal.AnimeEntry{
		Episode:        a.EpisodesWatched,
		Comments:       a.Notes,
		TimesRewatched: a.TimesRewatched,
	}
	// Status
	status, err := toMALStatus(a.Status)
	if err != nil {
		return mal.AnimeEntry{}, err
	}
	a.Status = status

	// rating
	if a.Rating != "" {
		f, err := strconv.ParseFloat(a.Rating, 64)
		if err == nil {
			f = math.Ceil(f * 2)
			score := int(f)
			e.Score = score
		}
	}
	if a.Rewatching {
		e.EnableRewatching = 1
	}
	return e, nil
}

func toMALEntries(anime []Anime) []mal.AnimeEntry {
	var malEntries []mal.AnimeEntry
	for _, a := range anime {
		e, err := toMALEntry(a)
		if err != nil {
			log.Printf("Discarded to MAL entry %v", err)
			continue
		}
		malEntries = append(malEntries, e)
	}
	return malEntries
}

func printAnimeUpdate(a Anime, op string) {
	fmt.Printf("%v %7v \t%v ", op, a.ID, a.Title)
	fmt.Printf("with values (")
	fmt.Printf("Status: %v, ", a.Status)
	fmt.Printf("EpisodesWatched: %v, ", a.EpisodesWatched)
	fmt.Printf("Rating: %v, ", a.Rating)
	fmt.Printf("Rewatching: %v, ", a.Rewatching)
	fmt.Printf("TimesRewatched: %v, ", a.TimesRewatched)
	fmt.Printf("Notes: %v)\n", a.Notes)
}
