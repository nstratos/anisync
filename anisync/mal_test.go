package anisync_test

import (
	"fmt"

	"github.com/nstratos/go-myanimelist/mal"
)

type MALClientStub struct {
	client *mal.Client
}

func NewMALClientStub(malClient *mal.Client, malAgent string) *MALClientStub {
	c := &MALClientStub{client: mal.NewClient()}
	c.client.SetUserAgent(malAgent)
	return c
}

func (c *MALClientStub) Verify(username, password string) error {
	switch {
	case username == "TestUsername" && password == "TestPassword":
		return nil
	default:
		return fmt.Errorf("wrong password")
	}
}
