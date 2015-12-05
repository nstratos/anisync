package anisync

import (
	"reflect"
	"testing"

	"github.com/nstratos/go-myanimelist/mal"
)

var toMALEntryTests = []struct {
	in  Anime
	out mal.AnimeEntry
}{
	{
		Anime{Status: StatusCurrentlyWatching},
		mal.AnimeEntry{Status: "1"},
	},
	{
		Anime{
			Status:          StatusOnHold,
			EpisodesWatched: 5,
		},
		mal.AnimeEntry{
			Status:  "3",
			Episode: 5,
		},
	},
	{
		Anime{
			Status: StatusOnHold,
			Rating: "4.5",
		},
		mal.AnimeEntry{
			Status: "3",
			Score:  9,
		},
	},
}

func Test_toMALEntry(t *testing.T) {
	for _, tt := range toMALEntryTests {
		got, err := toMALEntry(tt.in)
		if err != nil {
			t.Errorf("toMALEntry(%+v) returned error %v", tt.in, err)
		}
		if want := tt.out; !reflect.DeepEqual(got, want) {
			t.Errorf("toMALEntry(%+v) => \n%+v, want \n%+v", tt.in, got, want)
		}
	}
}

func Test_toMALEntry_noStatus(t *testing.T) {
	in := Anime{}
	_, err := toMALEntry(in)
	if err == nil {
		t.Errorf("toMALEntry with no status expected to return err")
	}
}
