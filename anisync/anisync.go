package anisync

import (
	"net/http"
	"sort"
	"time"

	"github.com/nstratos/go-myanimelist/mal"
)

type Client struct {
	resources Resources
}

func (c *Client) Resources() Resources { return c.resources }

func NewClient(resources Resources) *Client {
	return &Client{resources: resources}
}

func (c *Client) VerifyMALCredentials(username, password string) (*mal.User, *http.Response, error) {
	u, resp, err := c.resources.VerifyCredentials(username, password)
	if resp == nil {
		return u, nil, err
	}
	return u, resp.Response, err
}

type Anime struct {
	ID              int
	Status          Status
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
	sort.Sort(ByID(anime))
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
