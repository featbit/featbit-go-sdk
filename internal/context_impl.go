package internal

import (
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"strings"
)

const (
	streamingPath = "/streaming"
	eventPath     = "/api/public/insight/track"
)

type SDKContext struct {
	envSecret    string
	streamingUrl string
	eventUrl     string
	network      Network
}

func FromConfig(envSecret string, streamingUrl string, eventUrl string, factory NetworkFactory) (*SDKContext, error) {
	var err error
	ctx := &SDKContext{envSecret: envSecret, streamingUrl: streamingUrl, eventUrl: eventUrl}
	if factory != nil {
		ctx.network, err = factory.CreateNetwork(ctx)
	}
	return ctx, err
}

func (c *SDKContext) GetEnvSecret() string {
	return c.envSecret
}

func (c *SDKContext) GetStreamingUri() string {
	url := strings.TrimRight(c.streamingUrl, "/")
	return strings.Join([]string{url, streamingPath}, "")
}

func (c *SDKContext) GetEventUri() string {
	url := strings.TrimRight(c.eventUrl, "/")
	return strings.Join([]string{url, eventPath}, "")
}

func (c *SDKContext) GetNetwork() Network {
	return c.network
}
