package fbnetwork

import (
	"github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/gorilla/websocket"
	"net/http"
)

type NetworkConfigImpl struct {
	DefaultHeaders         http.Header
	HTTPClientFactory      func() *http.Client
	WebsocketClientFactory func() *websocket.Dialer
}

func (n NetworkConfigImpl) GetHeaders(headers map[string]string) http.Header {
	res := make(http.Header, len(n.DefaultHeaders))
	for k, v := range n.DefaultHeaders {
		res[k] = v
	}
	for k, v := range headers {
		res.Set(k, v)
	}
	return res
}

func (n NetworkConfigImpl) GetHTTPClient() interfaces.NetworkClient {
	if n.HTTPClientFactory == nil {
		client := *http.DefaultClient
		return &client
	}
	return n.HTTPClientFactory()
}

func (n NetworkConfigImpl) GetWebsocketClient() interfaces.NetworkClient {
	if n.WebsocketClientFactory == nil {
		dialer := *websocket.DefaultDialer
		return &dialer
	}
	return n.WebsocketClientFactory()
}
