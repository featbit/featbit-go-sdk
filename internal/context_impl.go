package internal

import (
	. "github.com/featbit/featbit-go-sdk/v2/interfaces"
	"log/slog"
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
	logger       *slog.Logger
}

func FromConfig(envSecret string, streamingUrl string, eventUrl string, factory NetworkFactory, loggers ...*slog.Logger) (*SDKContext, error) {
	var logger *slog.Logger
	if len(loggers) > 0 {
		logger = loggers[0]
	}
	var err error
	ctx := &SDKContext{envSecret: envSecret, streamingUrl: streamingUrl, eventUrl: eventUrl, logger: logger}
	if factory != nil {
		ctx.network, err = factory.CreateNetwork(ctx)
	}
	return ctx, err
}

func (c *SDKContext) GetLogger() *slog.Logger {
	return c.logger
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
