package anisync

import (
	"net/http"

	"github.com/nstratos/go-hummingbird/hb"
)

// HBClient is a Hummingbird client that contains implementations for all the
// operations that we need from the Hummingbird.me API.
type HBClient struct {
	client *hb.Client
}

// NewHBClient creates a new Hummingbird client.
func NewHBClient(client *hb.Client) *HBClient {
	return &HBClient{client: client}
}

func (c *HBClient) HBAnimeList(username string) ([]hb.LibraryEntry, *http.Response, error) {
	return c.client.User.Library(username, "")
}
