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

// Resources is an interface of all the operations we need from the external
// resources (MyAnimeList.net and Hummingbird.me APIs). It can be injected in
// anisync.Client which makes it easier to mock these operations during
// testing.
type Resources interface {
	MAL
	HB
}

// NewResources returns a Resources implementation that consists of a MALClient
// and a HBClient which are implementations of their respective MAL and HB
// interfaces. That implementation can be injected in anisync.Client using
// anisync.NewClient which is useful for testing. Alternatively, a new
// anisync.Client can also be created by anisync.NewDefaultClient which uses
// this function internally. In the typical case NewDefaultClient will be used
// in the program while the combination of NewResources and NewClient will be
// used for testing.
func NewResources(malClient *mal.Client, malAgent string, hbClient *hb.Client) Resources {
	return struct {
		*MALClient
		*HBClient
	}{
		NewMALClient(malClient, malAgent),
		NewHBClient(hbClient),
	}
}

// MAL is an interface describing all the operations that we need from the
// MyAnimeList.net API.
type MAL interface {
	Verify(username, password string) error
	MyAnimeList(username string) (*mal.AnimeList, *mal.Response, error)
}

// HB is an interface describing all the operations that we need from the
// Hummingbird.me API.
type HB interface {
}

// MALClient is a MyAnimeList client that contains implementations for all the
// operations that we need from the MyAnimeList.net API.
type MALClient struct {
	client *mal.Client
}

// NewMALClient creates a new MyAnimeList client that uses malAgent as user
// agent to communicate with the MyAnimeList.net API.
func NewMALClient(client *mal.Client, malAgent string) *MALClient {
	c := &MALClient{client: mal.NewClient()}
	c.client.SetUserAgent(malAgent)
	return c
}

func (c *MALClient) Verify(username, password string) error {
	c.client.SetCredentials(username, password)
	_, _, err := c.client.Account.Verify()
	return err
}

// MyAnimeList returns the anime list of a user.
func (c *MALClient) MyAnimeList(username string) (*mal.AnimeList, *mal.Response, error) {
	return c.client.Anime.List(username)
}

// HBClient is a Hummingbird client that contains implementations for all the
// operations that we need from the Hummingbird.met API.
type HBClient struct {
	client *hb.Client
}

// NewHBClient creates a new Hummingbird client.
func NewHBClient(client *hb.Client) *HBClient {
	return &HBClient{client: hb.NewClient(nil)}
}

type Client struct {
	mal       *mal.Client
	hb        *hb.Client
	resources Resources

	Anime *AnimeService
}

func NewDefaultClient(malAgent string) *Client {
	c := &Client{mal: mal.NewClient(), hb: hb.NewClient(nil)}
	c.mal.SetUserAgent(malAgent)
	c.Anime = &AnimeService{client: c}
	c.resources = NewResources(mal.NewClient(), malAgent, hb.NewClient(nil))
	return c
}

func NewClient(resources Resources) *Client {
	return &Client{resources: resources}
}

func (c *Client) VerifyMALCredentials(username, password string) error {
	return c.resources.Verify(username, password)
}

type AnimeService struct {
	client *Client
}

func (c *Client) GetMyAnimeList(username string) ([]Anime, *http.Response, error) {
	list, resp, err := c.resources.MyAnimeList(username)
	if err != nil {
		return nil, resp.Response, err
	}
	anime := fromMALEntries(*list)
	sort.Sort(ByID(anime))
	return anime, resp.Response, nil
}

//func (s *AnimeService) ListMAL(username string) ([]Anime, *http.Response, error) {
//	list, resp, err := s.client.mal.Anime.List(username)
//	if err != nil {
//		return nil, resp.Response, err
//	}
//	anime := fromMALEntries(*list)
//	sort.Sort(ByID(anime))
//	return anime, resp.Response, nil
//}

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

func fromMALMyLastUpdated(updated string) (*time.Time, error) {
	i, err := strconv.ParseInt(updated, 10, 64)
	if err != nil {
		return nil, err
	}
	t := time.Unix(i, 0).UTC()
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
