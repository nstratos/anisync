package anisync_test

import "github.com/nstratos/go-myanimelist/mal"

type MALClientStub struct {
	client *mal.Client
}

func NewMALClientStub(malClient *mal.Client, malAgent string) *MALClientStub {
	c := &MALClientStub{client: mal.NewClient(nil)}
	c.client.SetUserAgent(malAgent)
	return c
}
