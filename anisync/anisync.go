package anisync

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/nstratos/go-hummingbird/hb"
	"github.com/nstratos/go-myanimelist/mal"
)

const (
	StatusCurrentlyWatching = "currently-watching"
	StatusPlanToWatch       = "plan-to-watch"
	StatusCompleted         = "completed"
	StatusOnHold            = "on-hold"
	StatusDropped           = "dropped"
)

type Client struct {
	mal *mal.Client
	hb  *hb.Client

	Anime *AnimeService
}

func NewClient(malAgent string) *Client {
	c := &Client{mal: mal.NewClient(), hb: hb.NewClient(nil)}
	c.mal.SetUserAgent(malAgent)
	c.Anime = &AnimeService{client: c}
	return c
}

func (c *Client) VerifyMALCredentials(username, password string) error {
	c.mal.SetCredentials(username, password)
	_, _, err := c.mal.Account.Verify()
	return err
}

type AnimeService struct {
	client *Client
}

func (s *AnimeService) ListMAL(username string) ([]Anime, *http.Response, error) {
	list, resp, err := s.client.mal.Anime.List(username)
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
		a := fromMALEntry(mala)
		anime = append(anime, a)
	}
	return anime
}

func fromMALEntry(mala mal.Anime) Anime {
	a := Anime{
		ID:              mala.SeriesAnimeDBID,
		Title:           mala.SeriesTitle,
		EpisodesWatched: mala.MyWatchedEpisodes,
		Status:          fromMALStatus(mala.MyStatus),
		TimesRewatched:  mala.MyRewatchingEp,
		Image:           mala.SeriesImage,
		//Notes:           mala.Comments, // MAL API does not send the comments.
	}
	// LastUpdated
	lastUpdated, err := fromMALMyLastUpdated(mala.MyLastUpdated)
	if err != nil {
		log.Println("Could not parse mal time:", err)
	}
	a.LastUpdated = lastUpdated
	// Rating
	score := float64(mala.MyScore) / 2
	a.Rating = fmt.Sprintf("%.1f", score)
	// Rewatching
	if mala.MyRewatching == "1" {
		a.Rewatching = true
	}
	return a
}

func toMALEntry(a Anime) mal.AnimeEntry {
	e := mal.AnimeEntry{
		Episode:        a.EpisodesWatched,
		Status:         toMALStatus(a.Status),
		Comments:       a.Notes,
		TimesRewatched: a.TimesRewatched,
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

func toMALEntries(anime []Anime) []mal.AnimeEntry {
	var malEntries []mal.AnimeEntry
	for _, a := range anime {
		e := toMALEntry(a)
		malEntries = append(malEntries, e)
	}
	return malEntries
}

func fromMALMyLastUpdated(updated string) (*time.Time, error) {
	i, err := strconv.ParseInt(updated, 10, 64)
	if err != nil {
		return nil, err
	}
	t := time.Unix(i, 0)
	return &t, nil
}

func (s *AnimeService) ListHB(username string) ([]Anime, *http.Response, error) {
	entries, resp, err := s.client.hb.User.Library(username, "")
	if err != nil {
		return nil, resp, err
	}
	anime := fromHBEntries(entries)
	sort.Sort(ByID(anime))
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
		ID:              hbe.Anime.MALID,
		Title:           hbe.Anime.Title,
		EpisodesWatched: hbe.EpisodesWatched,
		Status:          hbe.Status,
		LastUpdated:     hbe.UpdatedAt,
		Notes:           hbe.Notes,
		TimesRewatched:  hbe.RewatchedTimes,
		Rewatching:      hbe.Rewatching,
		Image:           hbe.Anime.CoverImage,
	}
	// rating
	if hbe.Rating != nil {
		if hbe.Rating.Type == "advanced" {
			a.Rating = hbe.Rating.Value
		}
	}
	return a
}

func FindByID(anime []Anime, id int) *Anime {
	i := sort.Search(len(anime), func(i int) bool { return anime[i].ID >= id })
	if i < len(anime) && anime[i].ID == id {
		return &anime[i]
	}
	return nil
}

func FindByTitle(anime []Anime, title string) *Anime {
	i := sort.Search(len(anime), func(i int) bool { return anime[i].Title >= title })
	if i < len(anime) && anime[i].Title == title {
		return &anime[i]
	}
	return nil
}

//1/watching, 2/completed, 3/onhold, 4/dropped, 6/plantowatch
func fromMALStatus(status int) string {
	switch status {
	case 1:
		return StatusCurrentlyWatching
	case 2:
		return StatusCompleted
	case 3:
		return StatusOnHold
	case 4:
		return StatusDropped
	case 6:
		return StatusPlanToWatch
	default:
		return ""
	}
}

func toMALStatus(status string) string {
	switch status {
	case StatusCurrentlyWatching:
		return "1"
	case StatusCompleted:
		return "2"
	case StatusOnHold:
		return "3"
	case StatusDropped:
		return "4"
	case StatusPlanToWatch:
		return "6"
	default:
		return "1"
	}
}

type Anime struct {
	ID              int
	Status          string
	Title           string
	EpisodesWatched int
	LastUpdated     *time.Time
	Rating          string
	Notes           string
	TimesRewatched  int
	Rewatching      bool
	Image           string
}

type ByID []Anime

func (a ByID) Len() int           { return len(a) }
func (a ByID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByID) Less(i, j int) bool { return a[i].ID < a[j].ID }

type ByTitle []Anime

func (a ByTitle) Len() int           { return len(a) }
func (a ByTitle) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTitle) Less(i, j int) bool { return a[i].Title < a[j].Title }

// Diff represents the difference of two anime lists (left and right). It
// contains the orignal lists, the missing anime, the anime that need to
// be updated and the ones that are up to date. It is assuming that right
// list is larger than left list. Typically the left list will be the
// MyAnimeList and the right list will be the Hummingbird list.
type Diff struct {
	Left       []Anime
	Right      []Anime
	Missing    []Anime
	NeedUpdate []AniDiff
	UpToDate   []Anime
}

// Compare compares two anime lists and returns the difference containing
// the orignal lists, the missing anime, the anime that need to be updated
// and the ones that are up to date. It is assuming that right list is
// larger than left list. Typically the left list will be the MyAnimeList
// and the right list will be the Hummingbird list.
func Compare(left, right []Anime) *Diff {
	diff := &Diff{Left: left, Right: right}
	var (
		missing    []Anime
		needUpdate []AniDiff
		upToDate   []Anime
	)
	for _, a := range right {
		found := FindByID(left, a.ID)
		if found != nil {
			//fmt.Printf("found: %+v\n", found)
			needsUpdate, diff := compare(*found, a)
			if needsUpdate {
				needUpdate = append(needUpdate, diff)
			} else {
				upToDate = append(upToDate, a)
			}
		} else {
			missing = append(missing, a)
		}
	}
	diff.Missing = missing
	diff.NeedUpdate = needUpdate
	diff.UpToDate = upToDate
	return diff
}

type AniDiff struct {
	Anime           Anime
	Status          *Status
	EpisodesWatched *EpisodesWatched
	Rating          *Rating
	Rewatching      *Rewatching
	LastUpdated     *LastUpdated
}

type Status struct {
	Got  string
	Want string
}

type EpisodesWatched struct {
	Got  int
	Want int
}

type Rating struct {
	Got  string
	Want string
}

type Rewatching struct {
	Got  bool
	Want bool
}

type LastUpdated struct {
	Got  time.Time
	Want time.Time
}

func compare(left, right Anime) (bool, AniDiff) {
	needsUpdate := false
	diff := AniDiff{Anime: right}
	if got, want := left.Status, right.Status; got != want {
		diff.Status = &Status{got, want}
		// fmt.Printf("->Status got %v, want %v\n", got, want)
		needsUpdate = true
	}
	if got, want := left.EpisodesWatched, right.EpisodesWatched; got != want {
		//fmt.Printf("->EpisodesWatched got %v, want %v\n", got, want)
		diff.EpisodesWatched = &EpisodesWatched{got, want}
		needsUpdate = true
	}
	if got, want := left.Rating, right.Rating; got != want {
		//fmt.Printf("->Rating got %v, want %v\n", got, want)
		diff.Rating = &Rating{got, want}
		needsUpdate = true
	}
	if got, want := left.Rewatching, right.Rewatching; got != want {
		//fmt.Printf("->Rewatching got %v, want %v\n", got, want)
		diff.Rewatching = &Rewatching{got, want}
		needsUpdate = true
	}
	if left.LastUpdated != nil && right.LastUpdated != nil {
		// MAL API does not return comments so we cannot compare with notes.
		// It does not return times rewatched either. The only thing we can do
		// is compare the last updates. The problem is that MAL does not
		// always change last update when a change happens.
		c := compareLastUpdate(left, right)
		if got, want := left.LastUpdated, right.LastUpdated; c == -1 {
			diff.LastUpdated = &LastUpdated{*got, *want}
			needsUpdate = true
		}
	}
	return needsUpdate, diff
}

//
// if 0 it means anime are equal.
// if -1 it means right has more than left.
// if 1 it means left has more than right.
func compareLastUpdate(left, right Anime) int {
	if left.LastUpdated.Before(*right.LastUpdated) {
		return -1
	}
	if left.LastUpdated.After(*right.LastUpdated) {
		return 1
	}
	if left.LastUpdated.Equal(*right.LastUpdated) {
		return 0
	}
	return 0
}

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
	e := toMALEntry(a)
	printAnimeUpdate(a, "updating")
	_, err := s.client.mal.Anime.Update(a.ID, e)
	if err != nil {
		return err
	}
	return nil
}

func (s *AnimeService) AddMALAnime(a Anime) error {
	e := toMALEntry(a)
	printAnimeUpdate(a, "adding")
	_, err := s.client.mal.Anime.Add(a.ID, e)
	if err != nil {
		return err
	}
	return nil
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
