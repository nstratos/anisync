package anisync

import (
	"testing"

	"github.com/nstratos/go-myanimelist/mal"
)

var fromMALStatusTests = []struct {
	in  mal.Status
	out Status
}{
	{mal.Current, Current},
	{mal.Completed, Completed},
	{mal.OnHold, OnHold},
	{mal.Dropped, Dropped},
	{mal.Planned, Planned},
	{0, Unknown},
}

func TestFromMALStatus(t *testing.T) {
	for _, tt := range fromMALStatusTests {
		got := FromMALStatus(tt.in)
		if want := tt.out; got != want {
			t.Errorf("fromMALStatus(%q) => %q, want %q", tt.in, got, want)
		}
	}
}

func TestFromMALStatus_invalidStatus(t *testing.T) {
	var in mal.Status
	got := FromMALStatus(in)
	if want := Unknown; got != want {
		t.Errorf("fromMALStatus(%q) = %v, want %v", in, got, want)
	}
}

var toMALStatusTests = []struct {
	in  Status
	out mal.Status
}{
	{Current, mal.Current},
	{Completed, mal.Completed},
	{OnHold, mal.OnHold},
	{Dropped, mal.Dropped},
	{Planned, mal.Planned},
	{Unknown, 0},
}

func TestToMALStatus(t *testing.T) {
	for _, tt := range toMALStatusTests {
		got := toMALStatus(tt.in)
		if want := tt.out; got != want {
			t.Errorf("toMALStatus(%q) => %q, want %q", tt.in, got, want)
		}
	}
}

//func TestToMALStatus_invalidStatus(t *testing.T) {
//	var in string
//	_, err := toMALStatus(in)
//	if err == nil {
//		t.Errorf("toMALStatus(%q) expected to return err", in)
//	}
//}
