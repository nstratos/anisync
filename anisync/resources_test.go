package anisync_test

import (
	"reflect"
	"testing"

	"bitbucket.org/nstratos/anisync/anisync"
	"github.com/nstratos/go-kitsu/kitsu"
	"github.com/nstratos/go-myanimelist/mal"
)

func TestNewResources(t *testing.T) {
	c := anisync.NewResources(mal.NewClient(), kitsu.NewClient(nil))
	want := struct {
		*anisync.MALClient
		*anisync.HBClient
		*anisync.KitsuClient
	}{
		anisync.NewMALClient(mal.NewClient()),
		nil,
		anisync.NewKitsuClient(kitsu.NewClient(nil)),
	}
	if got := c; !reflect.DeepEqual(got, want) {
		t.Errorf("NewResources returned %+v, want %+v", got, want)
	}
}
