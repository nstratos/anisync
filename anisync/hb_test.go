package anisync_test

import "github.com/nstratos/go-hummingbird/hb"

type HBClientStub struct {
	client *hb.Client
}

func NewHBClientStub(hbClient *hb.Client) *HBClientStub {
	return &HBClientStub{client: hb.NewClient(nil)}
}
