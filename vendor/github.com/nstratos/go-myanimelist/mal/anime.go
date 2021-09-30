package mal

import (
	"encoding/xml"
	"fmt"
	"net/url"
)

// AnimeEntry represents the values that an anime will have on the list when
// added or updated. Status is required.
type AnimeEntry struct {
	XMLName            xml.Name `xml:"entry"`
	Episode            int      `xml:"episode"`
	Status             Status   `xml:"status,omitempty"` // Use the package constants: mal.Current, mal.Completed, etc.
	Score              int      `xml:"score"`
	DownloadedEpisodes int      `xml:"downloaded_episodes,omitempty"`
	StorageType        int      `xml:"storage_type,omitempty"`
	StorageValue       float64  `xml:"storage_value,omitempty"`
	TimesRewatched     int      `xml:"times_rewatched"`
	RewatchValue       int      `xml:"rewatch_value,omitempty"`
	DateStart          string   `xml:"date_start,omitempty"`  // mmddyyyy
	DateFinish         string   `xml:"date_finish,omitempty"` // mmddyyyy
	Priority           int      `xml:"priority,omitempty"`
	EnableDiscussion   int      `xml:"enable_discussion,omitempty"` // 1=enable, 0=disable
	EnableRewatching   int      `xml:"enable_rewatching"`           // 1=enable, 0=disable
	Comments           string   `xml:"comments"`
	FansubGroup        string   `xml:"fansub_group,omitempty"`
	Tags               string   `xml:"tags,omitempty"` // comma separated
}

// AnimeService handles communication with the Anime List methods of the
// MyAnimeList API.
//
// MyAnimeList API docs: http://myanimelist.net/modules.php?go=api
type AnimeService struct {
	client         *Client
	AddEndpoint    *url.URL
	UpdateEndpoint *url.URL
	DeleteEndpoint *url.URL
	SearchEndpoint *url.URL
	ListEndpoint   *url.URL
}

// Add allows an authenticated user to add an anime to their anime list.
func (s *AnimeService) Add(animeID int, entry AnimeEntry) (*Response, error) {

	return s.client.post(s.AddEndpoint.String(), animeID, entry, true)
}

// Update allows an authenticated user to update an anime on their anime list.
//
// Note: MyAnimeList.net updates the MyLastUpdated value of an Anime only if it
// receives an Episode update change. For example:
//
//    Updating Episode 0 -> 1 will update MyLastUpdated
//    Updating Episode 0 -> 0 will not update MyLastUpdated
//    Updating Status  1 -> 2 will not update MyLastUpdated
//    Updating Rating  5 -> 8 will not update MyLastUpdated
//
// As a consequence, you might perform a number of updates on a certain anime
// that will not affect its MyLastUpdate unless one of the updates happens to
// change the episode. This behavior is important to know if your application
// performs updates and cares about when an anime was last updated.
func (s *AnimeService) Update(animeID int, entry AnimeEntry) (*Response, error) {

	return s.client.post(s.UpdateEndpoint.String(), animeID, entry, true)
}

// Delete allows an authenticated user to delete an anime from their anime list.
func (s *AnimeService) Delete(animeID int) (*Response, error) {

	return s.client.delete(s.DeleteEndpoint.String(), animeID, true)
}

// AnimeResult represents the result of an anime search.
type AnimeResult struct {
	Rows []AnimeRow `xml:"entry"`
}

// AnimeRow represents each row of an anime search result.
type AnimeRow struct {
	ID        int     `xml:"id"`
	Title     string  `xml:"title"`
	English   string  `xml:"english"`
	Synonyms  string  `xml:"synonyms"`
	Score     float64 `xml:"score"`
	Type      string  `xml:"type"`
	Status    string  `xml:"status"`
	StartDate string  `xml:"start_date"`
	EndDate   string  `xml:"end_date"`
	Synopsis  string  `xml:"synopsis"`
	Image     string  `xml:"image"`
	Episodes  int     `xml:"episodes"`
}

// Search allows an authenticated user to search anime titles. If nothing is
// found, it will return an ErrNoContent error.
func (s *AnimeService) Search(query string) (*AnimeResult, *Response, error) {

	v := s.SearchEndpoint.Query()
	v.Set("q", query)
	s.SearchEndpoint.RawQuery = v.Encode()

	result := new(AnimeResult)
	resp, err := s.client.get(s.SearchEndpoint.String(), result, true)
	if err != nil {
		return nil, resp, err
	}
	return result, resp, nil
}

// AnimeList represents the anime list of a user.
type AnimeList struct {
	MyInfo AnimeMyInfo `xml:"myinfo"`
	Anime  []Anime     `xml:"anime"`
	Error  string      `xml:"error"`
}

// AnimeMyInfo represents the user's info which contains stats about the anime
// that exist in their anime list. For example how many anime they have
// completed, how many anime they are currently watching etc. It is returned as
// part of their AnimeList.
type AnimeMyInfo struct {
	ID                int    `xml:"user_id"`
	Name              string `xml:"user_name"`
	Completed         int    `xml:"user_completed"`
	OnHold            int    `xml:"user_onhold"`
	Dropped           int    `xml:"user_dropped"`
	DaysSpentWatching string `xml:"user_days_spent_watching"`
	Watching          int    `xml:"user_watching"`
	PlanToWatch       int    `xml:"user_plantowatch"`
}

// Anime represents a MyAnimeList anime. The data of the anime are stored in
// the fields starting with Series. User specific data are stored in the fields
// starting with My. For example, the score the user has set for that anime is
// stored in the MyScore field.
type Anime struct {
	SeriesAnimeDBID   int    `xml:"series_animedb_id"`
	SeriesEpisodes    int    `xml:"series_episodes"`
	SeriesTitle       string `xml:"series_title"`
	SeriesSynonyms    string `xml:"series_synonyms"`
	SeriesType        int    `xml:"series_type"`
	SeriesStatus      int    `xml:"series_status"`
	SeriesStart       string `xml:"series_start"`
	SeriesEnd         string `xml:"series_end"`
	SeriesImage       string `xml:"series_image"`
	MyID              int    `xml:"my_id"`
	MyStartDate       string `xml:"my_start_date"`
	MyFinishDate      string `xml:"my_finish_date"`
	MyScore           int    `xml:"my_score"`
	MyStatus          Status `xml:"my_status"` // Use the package constants: mal.Current, mal.Completed, etc.
	MyRewatching      int    `xml:"my_rewatching"`
	MyRewatchingEp    int    `xml:"my_rewatching_ep"`
	MyLastUpdated     string `xml:"my_last_updated"`
	MyTags            string `xml:"my_tags"`
	MyWatchedEpisodes int    `xml:"my_watched_episodes"`
}

// List allows an authenticated user to receive the anime list of a user.
func (s *AnimeService) List(username string) (*AnimeList, *Response, error) {

	v := s.ListEndpoint.Query()
	v.Set("status", "all")
	v.Set("type", "anime")
	v.Set("u", username)
	s.ListEndpoint.RawQuery = v.Encode()

	list := new(AnimeList)
	resp, err := s.client.get(s.ListEndpoint.String(), list, false)
	if err != nil {
		return nil, resp, err
	}

	if list.Error != "" {
		return list, resp, fmt.Errorf("%v", list.Error)
	}

	return list, resp, nil
}
