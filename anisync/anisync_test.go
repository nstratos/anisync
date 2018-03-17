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

var (
	client *Client
)

func init() {
	resources := struct {
		*MALClientStub
		*HBClientStub
	}{
		NewMALClientStub(mal.NewClient()),
		NewHBClientStub(hb.NewClient(nil)),
	}
	client = NewClient(resources)
}

func (c *MALClientStub) VerifyCredentials(username, password string) (*mal.User, *mal.Response, error) {
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
	u, _, err := client.VerifyMALCredentials("TestUsername", "TestPassword")
	if err != nil {
		t.Errorf("VerifyMALCredentials with correct username and password returned err: %v", err)
	}
	if got, want := u.Username, "TestUsername"; got != want {
		t.Errorf("VerifyMALCredentials with correct username and password returned username %q, want %q\n", got, want)
	}
}

func TestClient_VerifyMALCredentials_wrongPassword(t *testing.T) {
	_, _, err := client.VerifyMALCredentials("TestUser", "WrongTestPassword")
	if err == nil {
		t.Error("VerifyMALCredentials with wrong password expected to return err")
	}
}

func TestClient_VerifyMALCredentials_noResponse(t *testing.T) {
	_, resp, err := client.VerifyMALCredentials("TestNoResponse", "")
	if err == nil {
		t.Error("VerifyMALCredentials with no response expected to return err")
	}
	if resp != nil {
		t.Error("VerifyMALCredentials with no response expected to return nil response")
	}
}

func TestNewDefaultClient(t *testing.T) {
	c := NewDefaultClient("TestUsername", "TestPassword")
	got := c.Resources()
	want := NewResources(mal.NewClient(mal.Auth("TestUsername", "TestPassword")), "", hb.NewClient(nil))

	if !reflect.DeepEqual(got, want) {
		t.Errorf("NewDefaultClient.Resources() = \n%#v, want \n%#v", got, want)
	}
}
