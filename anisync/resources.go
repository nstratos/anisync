package anisync

import (
	"net/http"

	"github.com/nstratos/go-hummingbird/hb"
	"github.com/nstratos/go-kitsu/kitsu"
	"github.com/nstratos/go-myanimelist/mal"
)

// Resources is an interface of all the operations we need from the external
// resources (MyAnimeList.net and Hummingbird.me APIs). It can be injected in
// anisync.Client which makes it easier to mock these operations during
// testing.
type Resources interface {
	MAL
	HB
	Kitsu
}

// NewResources returns a Resources implementation that consists of a MALClient
// and a HBClient which are implementations of their respective MAL and HB
// interfaces. That implementation can be injected in anisync.Client using
// anisync.NewClient which is useful for testing. Alternatively, a new
// anisync.Client can also be created by anisync.NewDefaultClient which uses
// this function internally. In the typical case NewDefaultClient will be used
// in the program while the combination of NewResources and NewClient will be
// used for testing.
func NewResources(malClient *mal.Client, kitsuClient *kitsu.Client) Resources {
	return struct {
		*MALClient
		*HBClient
		*KitsuClient
	}{
		NewMALClient(malClient),
		nil,
		NewKitsuClient(kitsuClient),
	}
}

// MAL is an interface describing all the operations that we need from the
// MyAnimeList.net API.
type MAL interface {
	VerifyCredentials(username, password string) (*mal.User, *mal.Response, error)
	MyAnimeList(username string) (*mal.AnimeList, *mal.Response, error)
	UpdateMALAnimeEntry(id int, entry mal.AnimeEntry) (*mal.Response, error)
	AddMALAnimeEntry(id int, entry mal.AnimeEntry) (*mal.Response, error)
}

// HB is an interface describing all the operations that we need from the
// Hummingbird.me API.
type HB interface {
	HBAnimeList(username string) ([]hb.LibraryEntry, *http.Response, error)
}

type Kitsu interface {
	KitsuAnimeList(userID string) ([]*kitsu.LibraryEntry, *kitsu.Response, error)
}
