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
		Anime{Status: Current},
		mal.AnimeEntry{Status: mal.Current},
	},
	{
		Anime{
			Status:          OnHold,
			EpisodesWatched: 5,
		},
		mal.AnimeEntry{
			Status:  mal.OnHold,
			Episode: 5,
		},
	},
	{
		Anime{
			Status: OnHold,
			Rating: "4.5",
		},
		mal.AnimeEntry{
			Status: mal.OnHold,
			Score:  9,
		},
	},
}

func Test_toMALEntry(t *testing.T) {
	for _, tt := range toMALEntryTests {
		got := toMALEntry(tt.in)
		if want := tt.out; !reflect.DeepEqual(got, want) {
			t.Errorf("toMALEntry(%+v) => \n%+v, want \n%+v", tt.in, got, want)
		}
	}
}

//func Test_toMALEntry_noStatus(t *testing.T) {
//	in := Anime{}
//	_, err := toMALEntry(in)
//	if err == nil {
//		t.Errorf("toMALEntry with no status expected to return err")
//	}
//}
