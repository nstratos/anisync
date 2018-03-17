package anisync_test

import "github.com/nstratos/go-myanimelist/mal"

type MALClientStub struct {
	client *mal.Client
}

func NewMALClientStub(malClient *mal.Client) *MALClientStub {
	return &MALClientStub{client: mal.NewClient()}
}
