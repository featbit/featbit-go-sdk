package factories

import (
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal/datasynchronization"
	"math"
	"time"
)

const defaultFirstRetryDelay = time.Second

type StreamingBuilder struct {
	firstRetryDelay time.Duration
	maxRetryTimes   int
}

func NewStreamingBuilder() *StreamingBuilder {
	return &StreamingBuilder{firstRetryDelay: defaultFirstRetryDelay, maxRetryTimes: math.MaxInt32}
}

func (s *StreamingBuilder) FirstRetryDelay(firstRetryDelay time.Duration) *StreamingBuilder {
	if firstRetryDelay <= 0 {
		s.firstRetryDelay = defaultFirstRetryDelay
	} else {
		s.firstRetryDelay = firstRetryDelay
	}
	return s
}

func (s *StreamingBuilder) MaxRetryTimes(maxRetryTimes int) *StreamingBuilder {
	if maxRetryTimes <= 0 {
		s.maxRetryTimes = math.MaxInt32
	} else {
		s.maxRetryTimes = maxRetryTimes
	}
	return s
}

func (s *StreamingBuilder) CreateDataSynchronizer(context Context, dataUpdater DataUpdater) DataSynchronizer {
	return datasynchronization.NewStreaming(context, dataUpdater, s.firstRetryDelay, s.maxRetryTimes)
}
