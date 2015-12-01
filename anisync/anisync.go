package anisync

import (
	"sort"
	"time"

	"github.com/nstratos/go-hummingbird/hb"
	"github.com/nstratos/go-myanimelist/mal"
)

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

//func (s *AnimeService) ListMAL(username string) ([]Anime, *http.Response, error) {
//	list, resp, err := s.client.mal.Anime.List(username)
//	if err != nil {
//		return nil, resp.Response, err
//	}
//	anime := fromMALEntries(*list)
//	sort.Sort(ByID(anime))
//	return anime, resp.Response, nil
//}

//func (s *AnimeService) ListHB(username string) ([]Anime, *http.Response, error) {
//	entries, resp, err := s.client.hb.User.Library(username, "")
//	if err != nil {
//		return nil, resp, err
//	}
//	anime := fromHBEntries(entries)
//	sort.Sort(ByID(anime))
//	return anime, resp, nil
//
//}

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

func FindByID(anime []Anime, id int) *Anime {
	i := sort.Search(len(anime), func(i int) bool { return anime[i].ID >= id })
	if i < len(anime) && anime[i].ID == id {
		return &anime[i]
	}
	return nil
}

type ByID []Anime

func (a ByID) Len() int           { return len(a) }
func (a ByID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByID) Less(i, j int) bool { return a[i].ID < a[j].ID }
