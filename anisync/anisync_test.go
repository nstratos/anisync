package anisync_test

import (
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

func TestClient_VerifyMALCredentials(t *testing.T) {
	err := client.VerifyMALCredentials("TestUsername", "TestPassword")
	if err != nil {
		t.Errorf("VerifyMALCredentials with correct username and password expected to return nil")
	}
}

func TestClient_VerifyMALCredentials_wrongPassword(t *testing.T) {
	err := client.VerifyMALCredentials("TestUser", "WrongTestPassword")
	if err == nil {
		t.Errorf("VerifyMALCredentials with wrong password expected to return err")
	}
}

func TestNewDefaultClient(t *testing.T) {
	c := NewDefaultClient(defaultUserAgent)

	got := NewResources(mal.NewClient(nil), defaultUserAgent, hb.NewClient(nil))

	if want := c.Resources(); !reflect.DeepEqual(got, want) {
		t.Errorf("NewDefaultClient.Resources() = %q, want %q", got, want)
	}
}
