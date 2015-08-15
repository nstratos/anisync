package anisync

import (
	"log"
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

func (s *AnimeService) ListMAL(username string) ([]Anime, error) {
	list, _, err := s.client.mal.Anime.List(username)
	if err != nil {
		return nil, err
	}
	anime := fromListMAL(*list)
	sort.Sort(ByID(anime))
	return anime, nil

}

func fromListMAL(malist mal.AnimeList) []Anime {
	var anime []Anime
	for _, mala := range malist.Anime {
		a := Anime{
			ID:              mala.SeriesAnimeDBID,
			Title:           mala.SeriesTitle,
			EpisodesWatched: mala.MyWatchedEpisodes,
			Status:          fromMALStatus(mala.MyStatus),
		}
		lastUpdated, err := fromMALMyLastUpdated(mala.MyLastUpdated)
		if err != nil {
			log.Println("Could not parse mal time:", err)
		}
		a.LastUpdated = lastUpdated
		anime = append(anime, a)
	}
	return anime
}

func fromMALMyLastUpdated(updated string) (*time.Time, error) {
	i, err := strconv.ParseInt(updated, 10, 64)
	if err != nil {
		return nil, err
	}
	t := time.Unix(i, 0)
	return &t, nil
}

func (s *AnimeService) ListHB(username string) ([]Anime, error) {
	list, _, err := s.client.hb.User.Library(username, "")
	if err != nil {
		return nil, err
	}
	anime := fromListHB(list)
	sort.Sort(ByID(anime))
	return anime, nil

}

func fromListHB(list []hb.LibraryEntry) []Anime {
	var anime []Anime
	for _, hba := range list {
		a := Anime{
			ID:              hba.Anime.MALID,
			Title:           hba.Anime.Title,
			EpisodesWatched: hba.EpisodesWatched,
			Status:          hba.Status,
			LastUpdated:     hba.UpdatedAt,
		}
		anime = append(anime, a)
	}
	return anime
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

type Anime struct {
	ID              int
	Status          string
	Title           string
	EpisodesWatched int
	LastUpdated     *time.Time
}

type ByID []Anime

func (a ByID) Len() int           { return len(a) }
func (a ByID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByID) Less(i, j int) bool { return a[i].ID < a[j].ID }

type ByTitle []Anime

func (a ByTitle) Len() int           { return len(a) }
func (a ByTitle) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTitle) Less(i, j int) bool { return a[i].Title < a[j].Title }

// assuming right is larger than left.
// assuming right will update left.
type Diff struct {
	Left       []Anime
	Right      []Anime
	Missing    []Anime
	NeedUpdate []Anime
	UpToDate   []Anime
	Equal      []Anime
}

// Compare compares two anime lists. It is assuming right is larger than left.
func Compare(left, right []Anime) *Diff {
	diff := &Diff{Left: left, Right: right}
	var (
		missing    []Anime
		needUpdate []Anime
		upToDate   []Anime
		equal      []Anime
	)
	for _, a := range right {
		found := FindByID(left, a.ID)
		if found != nil {
			c := compare(*found, a)
			switch c {
			case 0:
				// equal, nothing to do
				equal = append(equal, a)
			case -1:
				// update for mal
				needUpdate = append(needUpdate, a)
			case 1:
				// update for hb
				upToDate = append(upToDate, a)
			}
		} else {
			missing = append(missing, a)
		}
	}
	diff.Missing = missing
	diff.NeedUpdate = needUpdate
	diff.UpToDate = upToDate
	diff.Equal = equal
	return diff
}

//
// if = 0 it means anime are equal.
// if < 0 it means right has more than left.
// if > 0 it means left has more than right.
func compare(left, right Anime) int {
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
