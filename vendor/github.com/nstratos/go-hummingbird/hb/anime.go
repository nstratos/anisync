package hb

import (
	"fmt"
	"net/http"
)

// Anime represents a hummingbird anime object.
type Anime struct {
	ID              int     `json:"id,omitempty"`
	MALID           int     `json:"mal_id,omitempty"`
	Slug            string  `json:"slug,omitempty"`
	Status          string  `json:"status,omitempty"`
	URL             string  `json:"url,omitempty"`
	Title           string  `json:"title,omitempty"`
	AlternateTitle  string  `json:"alternate_title,omitempty"`
	EpisodeCount    int     `json:"episode_count,omitempty"`
	EpisodeLength   int     `json:"episode_length,omitempty"`
	CoverImage      string  `json:"cover_image,omitempty"`
	Synopsis        string  `json:"synopsis,omitempty"`
	ShowType        string  `json:"show_type,omitempty"`
	StartedAiring   string  `json:"started_airing,omitempty"`
	FinishedAiring  string  `json:"finished_airing,omitempty"`
	CommunityRating float64 `json:"community_rating,omitempty"`
	AgeRating       string  `json:"age_rating,omitempty"`
	Genres          []Genre `json:"genres,omitempty"`
	FavID           int     `json:"fav_id,omitempty"`   // When requesting user favorite anime.
	FavRank         int     `json:"fav_rank,omitempty"` // When requesting user favorite anime.
}

// Genre represents the genre of an anime.
type Genre struct {
	Name string
}

// AnimeService handles communication with the anime methods of
// the Hummingbird API.
//
// Hummingbird API docs:
// https://github.com/hummingbird-me/hummingbird/wiki/API-v1-Methods#anime
type AnimeService struct {
	client *Client
}

// Get returns anime metadata based on ID which can be either the anime ID or
// a slug. An optional parameter about the title language preference can be
// used which can be one of: "canonical", "english", "romanized".
// If omitted, "canonical" will be used.
//
// Does not require authentication.
func (s *AnimeService) Get(animeID, titleLangPref string) (*Anime, *http.Response, error) {
	urlStr := fmt.Sprintf("api/v1/anime/%s", animeID)

	req, err := s.client.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, nil, err
	}

	if titleLangPref != "" {
		v := req.URL.Query()
		v.Set("title_language_preference", titleLangPref)
		req.URL.RawQuery = v.Encode()
	}

	anime := new(Anime)
	resp, err := s.client.Do(req, anime)
	if err != nil {
		return nil, resp, err
	}
	return anime, resp, nil
}

// Search allows searching anime by title. It returns an array of anime objects
// (5 max) without genres. It supports fuzzy search.
//
// Does not require authentication.
func (s *AnimeService) Search(query string) ([]Anime, *http.Response, error) {
	const urlStr = "api/v1/search/anime"

	req, err := s.client.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, nil, err
	}

	v := req.URL.Query()
	v.Set("query", query)
	req.URL.RawQuery = v.Encode()

	var anime []Anime
	resp, err := s.client.Do(req, &anime)
	if err != nil {
		return nil, resp, err
	}
	return anime, resp, nil
}
