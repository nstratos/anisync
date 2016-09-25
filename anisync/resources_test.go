package anisync_test

import (
	"reflect"
	"testing"

	"bitbucket.org/nstratos/anisync/anisync"
	"github.com/nstratos/go-hummingbird/hb"
	"github.com/nstratos/go-myanimelist/mal"
)

func TestNewResources(t *testing.T) {
	c := anisync.NewResources(mal.NewClient(nil), "", hb.NewClient(nil))
	want := struct {
		*anisync.MALClient
		*anisync.HBClient
	}{
		anisync.NewMALClient(mal.NewClient(nil), ""),
		anisync.NewHBClient(hb.NewClient(nil)),
	}
	if got := c; !reflect.DeepEqual(got, want) {
		t.Errorf("NewResources returned %+v, want %+v")
	}
}
