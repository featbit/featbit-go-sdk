package interfaces

import "net/http"

type NetworkClient interface{}

// Network encapsulates top-level HTTP or Websocket configuration that applies to all SDK components.
type Network interface {
	// GetHeaders creates the HTTP headers for the next requests
	GetHeaders(map[string]string) http.Header
	// GetHTTPClient get the HTTP Client
	GetHTTPClient() NetworkClient
	// GetWebsocketClient Get the Websocket Client
	GetWebsocketClient() NetworkClient
}

// NetworkFactory an interface of a factory that creates a Network.
type NetworkFactory interface {
	// CreateNetwork creates an implementation of Network
	CreateNetwork(BasicConfig) (Network, error)
}
