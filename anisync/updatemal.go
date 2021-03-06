package anisync

import (
	"math"
	"strconv"

	"github.com/nstratos/go-myanimelist/mal"
)

type Fail struct {
	Anime Anime
	Error error
}

type SyncResult struct {
	Adds        []AddSuccess
	AddFails    []AddFail
	Updates     []UpdateSuccess
	UpdateFails []UpdateFail
}

type AddSuccess struct {
	Anime Anime
}

type AddFail struct {
	Anime  Anime
	Error  error
	Reason string
}

func MakeAddFail(a Anime, err error) AddFail {
	return AddFail{Anime: a, Error: err, Reason: err.Error()}
}

type UpdateSuccess struct {
	AniDiff
}

type UpdateFail struct {
	AniDiff
	Error  error
	Reason string
}

func MakeUpdateFail(d AniDiff, err error) UpdateFail {
	return UpdateFail{AniDiff: d, Error: err, Reason: err.Error()}
}

func (c *Client) SyncMALAnime(diff Diff) *SyncResult {
	var adds []AddSuccess
	var addf []AddFail
	for _, a := range diff.Missing {
		err := c.AddMALAnime(a)
		if err != nil {
			addf = append(addf, MakeAddFail(a, err))
			continue
		}
		adds = append(adds, AddSuccess{Anime: a})
	}

	var upds []UpdateSuccess
	var updf []UpdateFail
	for _, d := range diff.NeedUpdate {
		err := c.UpdateMALAnime(d.Anime)
		if err != nil {
			updf = append(updf, MakeUpdateFail(d, err))
			continue
		}
		upds = append(upds, UpdateSuccess{AniDiff: d})
	}

	return &SyncResult{Adds: adds, AddFails: addf, Updates: upds, UpdateFails: updf}
}

func (c *Client) UpdateMALAnime(a Anime) error {
	e := toMALEntry(a)

	_, err := c.resources.UpdateMALAnimeEntry(a.ID, e)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) AddMALAnime(a Anime) error {
	e := toMALEntry(a)

	_, err := c.resources.AddMALAnimeEntry(a.ID, e)
	if err != nil {
		return err
	}
	return nil
}

func toMALEntry(a Anime) mal.AnimeEntry {
	e := mal.AnimeEntry{
		Episode:        a.EpisodesWatched,
		Comments:       a.Notes,
		TimesRewatched: a.TimesRewatched,
		Status:         toMALStatus(a.Status),
	}

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
	return e
}
