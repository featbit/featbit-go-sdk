package factories

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	. "github.com/featbit/featbit-go-sdk/interfaces"
	fbnetwork "github.com/featbit/featbit-go-sdk/internal/network"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultConnectTimeout = 5 * time.Second
	defaultReadTimeout    = 10 * time.Second
)

type NetworkBuilder struct {
	connectTimeout time.Duration
	readTimeout    time.Duration
	proxyURL       string
	caCert         string
	certFile       string
	keyFile        string
}

func NewNetworkBuilder() *NetworkBuilder {
	return &NetworkBuilder{connectTimeout: defaultConnectTimeout, readTimeout: defaultReadTimeout}
}

func (n *NetworkBuilder) ConnectTime(connectTimeout time.Duration) *NetworkBuilder {
	n.connectTimeout = connectTimeout
	return n
}

func (n *NetworkBuilder) ReadTime(readTimeout time.Duration) *NetworkBuilder {
	n.readTimeout = readTimeout
	return n
}

func (n *NetworkBuilder) ProxyURL(proxy string) *NetworkBuilder {
	n.proxyURL = proxy
	return n
}

func (n *NetworkBuilder) CaCert(caRoot string) *NetworkBuilder {
	n.caCert = caRoot
	return n
}

func (n *NetworkBuilder) CertFile(certFilePath string) *NetworkBuilder {
	n.certFile = certFilePath
	return n
}

func (n *NetworkBuilder) KeyFile(keyFilePath string) *NetworkBuilder {
	n.keyFile = keyFilePath
	return n
}

func (n *NetworkBuilder) CreateNetwork(config BasicConfig) (Network, error) {
	if n.connectTimeout <= 0 {
		n.connectTimeout = defaultConnectTimeout
	}
	if n.readTimeout <= 0 {
		n.readTimeout = defaultReadTimeout
	}
	// headers
	defaultHeaders := make(http.Header)
	defaultHeaders.Set("Authorization", config.GetEnvSecret())
	defaultHeaders.Set("User-Agent", "fb-go-server-sdk")
	defaultHeaders.Set("Content-Type", "application/json")
	// network config
	dialer := &net.Dialer{
		Timeout: n.connectTimeout,
	}
	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	var tlsConfig *tls.Config
	var caCerts *x509.CertPool
	var certs []tls.Certificate
	if n.caCert != "" {
		bytes, err := ioutil.ReadFile(n.certFile)
		if err != nil {
			return nil, err
		}
		caCerts, err = x509.SystemCertPool() // this returns a *copy* of the existing CA certs
		if err != nil {
			caCerts = x509.NewCertPool()
		}
		if caCerts.AppendCertsFromPEM(bytes) {
			return nil, fmt.Errorf("invalid CA certificate data")
		}
	}
	if n.keyFile != "" && n.certFile != "" {
		cliCert, err := tls.LoadX509KeyPair(n.certFile, n.keyFile)
		if err != nil {
			return nil, err
		}
		certs = []tls.Certificate{cliCert}
	} else if n.certFile != "" {
		bytes, err := ioutil.ReadFile(n.certFile)
		if err != nil {
			return nil, err
		}
		var cliCert tls.Certificate
		var block *pem.Block
		for {
			block, bytes = pem.Decode(bytes)
			if bytes == nil {
				break
			}
			if block.Type == "CERTIFICATE" {
				cliCert.Certificate = append(cliCert.Certificate, block.Bytes)
			}
		}
		if len(cliCert.Certificate) == 0 {
			return nil, fmt.Errorf("failed to find any PEM data in certificate input")
		}
		certs = []tls.Certificate{cliCert}
	}

	if caCerts != nil || len(certs) > 0 {
		tlsConfig = &tls.Config{InsecureSkipVerify: false}
		if caCerts != nil {
			tlsConfig.RootCAs = caCerts
		}
		if len(certs) > 0 {
			tlsConfig.Certificates = certs
		}
		transport.TLSClientConfig = tlsConfig
	}
	var proxy *url.URL
	if n.proxyURL != "" {
		var err error
		proxy, err = url.Parse(n.proxyURL)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxy)
	}

	httpClientFactory := func() *http.Client {
		return &http.Client{
			Timeout:   n.connectTimeout + n.readTimeout,
			Transport: transport,
		}
	}
	// websocket config
	websocketClientFactory := func() *websocket.Dialer {
		client := websocket.DefaultDialer
		client.NetDialContext = dialer.DialContext
		if tlsConfig != nil {
			// https://github.com/gorilla/websocket/issues/601
			client.TLSClientConfig = tlsConfig.Clone()
		}
		if proxy != nil {
			client.Proxy = http.ProxyURL(proxy)
		}
		return client
	}

	return fbnetwork.NetworkConfigImpl{
		DefaultHeaders:         defaultHeaders,
		HTTPClientFactory:      httpClientFactory,
		WebsocketClientFactory: websocketClientFactory,
	}, nil
}
