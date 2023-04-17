package factories

import (
	"fmt"
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal/datasynchronization"
	"github.com/gorilla/websocket"
	"math"
	"time"
)

const defaultFirstRetryDelay = time.Second

// StreamingBuilder factory to create default implementation of interfaces.DataSynchronizer
type StreamingBuilder struct {
	firstRetryDelay time.Duration
	maxRetryTimes   int
}

// NewStreamingBuilder creates an instance of StreamingBuilder
func NewStreamingBuilder() *StreamingBuilder {
	return &StreamingBuilder{firstRetryDelay: defaultFirstRetryDelay, maxRetryTimes: math.MaxInt32}
}

// FirstRetryDelay sets the time to wait for next retry if the last data synchronization failed
func (s *StreamingBuilder) FirstRetryDelay(firstRetryDelay time.Duration) *StreamingBuilder {
	if firstRetryDelay <= 0 {
		s.firstRetryDelay = defaultFirstRetryDelay
	} else {
		s.firstRetryDelay = firstRetryDelay
	}
	return s
}

// MaxRetryTimes sets the max retry times for the failed data synchronization
func (s *StreamingBuilder) MaxRetryTimes(maxRetryTimes int) *StreamingBuilder {
	if maxRetryTimes <= 0 {
		s.maxRetryTimes = math.MaxInt32
	} else {
		s.maxRetryTimes = maxRetryTimes
	}
	return s
}

// CreateDataSynchronizer creates an instance of interfaces.DataSynchronizer
func (s *StreamingBuilder) CreateDataSynchronizer(context Context, dataUpdater DataUpdater) (DataSynchronizer, error) {
	network := context.GetNetwork()
	_, ok := network.GetWebsocketClient().(*websocket.Dialer)
	if !ok {
		return nil, fmt.Errorf("non supported Websocket Client")
	}
	return datasynchronization.NewStreaming(context, dataUpdater, s.firstRetryDelay, s.maxRetryTimes), nil
}

type nullDataSynchronizerBuilder struct{}

func ExternalDataSynchronization() DataSynchronizerFactory {
	return &nullDataSynchronizerBuilder{}
}

func (n *nullDataSynchronizerBuilder) CreateDataSynchronizer(_ Context, dataUpdater DataUpdater) (DataSynchronizer, error) {
	return datasynchronization.NewNullDataSynchronizer(dataUpdater), nil
}
