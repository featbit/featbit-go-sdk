package interfaces

import "net/http"

type NetworkClient interface{}

// Network encapsulates top-level HTTP or Websocket configuration that applies to all SDK components.
type Network interface {
	GetHeaders(map[string]string) http.Header
	GetHTTPClient() NetworkClient
	GetWebsocketClient() NetworkClient
}

// NetworkFactory an interface of a factory that creates a Network.
type NetworkFactory interface {
	CreateNetwork(BasicConfig) (Network, error)
}
