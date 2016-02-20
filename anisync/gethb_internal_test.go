package anisync

import (
	"reflect"
	"testing"

	"github.com/nstratos/go-hummingbird/hb"
)

var fromHBEntryTests = []struct {
	in  hb.LibraryEntry
	out Anime
}{
	{hb.LibraryEntry{}, Anime{}},
	{hb.LibraryEntry{Anime: &hb.Anime{Title: "title"}}, Anime{Title: "title"}},
	{hb.LibraryEntry{Anime: &hb.Anime{MALID: 5}}, Anime{ID: 5}},
	{hb.LibraryEntry{Anime: &hb.Anime{MALID: 5, Title: "title"}}, Anime{ID: 5, Title: "title"}},
	{
		hb.LibraryEntry{
			Anime:  &hb.Anime{MALID: 5, Title: "title"},
			Rating: &hb.LibraryEntryRating{Type: "advanced", Value: "3.0"},
		},
		Anime{ID: 5, Title: "title", Rating: "3.0"},
	},
	{
		hb.LibraryEntry{
			Rating: &hb.LibraryEntryRating{Type: "advanced", Value: "0.0"},
		},
		Anime{Rating: "0.0"},
	},
	{
		hb.LibraryEntry{
			Rating: &hb.LibraryEntryRating{Type: "advanced", Value: ""},
		},
		Anime{Rating: ""},
	},
}

func Test_fromHBEntries(t *testing.T) {
	for _, tt := range fromHBEntryTests {
		got := fromHBEntry(tt.in)
		if want := tt.out; !reflect.DeepEqual(got, want) {
			t.Errorf("fromHBEntry() with input \n%#v \nhas output\n%#v \nbut want \n%#v", tt.in, got, want)
		}
	}
}
