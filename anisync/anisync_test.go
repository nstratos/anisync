package anisync_test

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	. "bitbucket.org/nstratos/anisync/anisync"

	"github.com/nstratos/go-hummingbird/hb"
	"github.com/nstratos/go-myanimelist/mal"
)

const defaultUserAgent = `
	Mozilla/5.0 (X11; Linux x86_64) 
	AppleWebKit/537.36 (KHTML, like Gecko) 
	Chrome/42.0.2311.90 Safari/537.36`

var (
	client *Client
)

func init() {
	resources := struct {
		*MALClientStub
		*HBClientStub
	}{
		NewMALClientStub(mal.NewClient(nil), defaultUserAgent),
		NewHBClientStub(hb.NewClient(nil)),
	}
	client = NewClient(resources)
}

func (c *MALClientStub) SetAndVerifyCredentials(username, password string) (*mal.User, *mal.Response, error) {
	switch {
	case username == "TestUsername" && password == "TestPassword":
		return &mal.User{Username: "TestUsername"}, &mal.Response{Response: &http.Response{}}, nil
	case username == "TestNoResponse":
		return nil, nil, fmt.Errorf("could not reach test myanimelist server")
	default:
		return nil, &mal.Response{Response: &http.Response{}}, fmt.Errorf("wrong password")
	}
}

func TestClient_VerifyMALCredentials(t *testing.T) {
	u, _, err := client.SetAndVerifyMALCredentials("TestUsername", "TestPassword")
	if err != nil {
		t.Errorf("VerifyMALCredentials with correct username and password returned err")
	}
	if got, want := u.Username, "TestUsername"; got != want {
		t.Errorf("VerifyMALCredentials with correct username and password returned username %q, want %q\n")
	}
}

func TestClient_VerifyMALCredentials_wrongPassword(t *testing.T) {
	_, _, err := client.SetAndVerifyMALCredentials("TestUser", "WrongTestPassword")
	if err == nil {
		t.Error("VerifyMALCredentials with wrong password expected to return err")
	}
}

func TestClient_VerifyMALCredentials_noResponse(t *testing.T) {
	_, resp, err := client.SetAndVerifyMALCredentials("TestNoResponse", "")
	if err == nil {
		t.Error("VerifyMALCredentials with no response expected to return err")
	}
	if resp != nil {
		t.Error("VerifyMALCredentials with no response expected to return nil response")
	}
}

func TestNewDefaultClient(t *testing.T) {
	c := NewDefaultClient(defaultUserAgent)

	got := NewResources(mal.NewClient(nil), defaultUserAgent, hb.NewClient(nil))

	if want := c.Resources(); !reflect.DeepEqual(got, want) {
		t.Errorf("NewDefaultClient.Resources() = %q, want %q", got, want)
	}
}
