package anisync

import "github.com/nstratos/go-myanimelist/mal"

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