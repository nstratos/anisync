package anisync

import "testing"

var fromMALStatusTests = []struct {
	in  int
	out string
}{
	{1, "currently-watching"},
	{2, "completed"},
	{3, "on-hold"},
	{4, "dropped"},
	{6, "plan-to-watch"},
}

func TestFromMALStatus(t *testing.T) {
	for _, tt := range fromMALStatusTests {
		got, _ := fromMALStatus(tt.in)
		if want := tt.out; got != want {
			t.Errorf("fromMALStatus(%q) => %q, want %q", tt.in, got, want)
		}
	}
}

func TestFromMALStatus_invalidStatus(t *testing.T) {
	var in int
	_, err := fromMALStatus(in)
	if err == nil {
		t.Errorf("fromMALStatus(%q) expected to return err", in)
	}
}

var toMALStatusTests = []struct {
	in  string
	out string
}{
	{"currently-watching", "1"},
	{"completed", "2"},
	{"on-hold", "3"},
	{"dropped", "4"},
	{"plan-to-watch", "6"},
}

func TestToMALStatus(t *testing.T) {
	for _, tt := range toMALStatusTests {
		got, _ := toMALStatus(tt.in)
		if want := tt.out; got != want {
			t.Errorf("toMALStatus(%q) => %q, want %q", tt.in, got, want)
		}
	}
}

func TestToMALStatus_invalidStatus(t *testing.T) {
	var in string
	_, err := toMALStatus(in)
	if err == nil {
		t.Errorf("toMALStatus(%q) expected to return err", in)
	}
}
