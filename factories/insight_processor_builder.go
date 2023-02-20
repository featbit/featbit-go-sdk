package factories

import (
	"fmt"
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal/insight"
	"net/http"
	"time"
)

const (
	DefaultCapacity      = 10000
	DefaultFlushInterval = time.Second
	DefaultRetryDelay    = 100 * time.Millisecond
	DefaultRetryTimes    = 1
)

// InsightProcessorBuilder factory to create default implementation of interfaces.InsightProcessor
type InsightProcessorBuilder struct {
	capacity      int
	flushInterval time.Duration
	retryInterval time.Duration
	maxRetryTimes int
}

// NewInsightProcessorBuilder creates an instance of InsightProcessorBuilder
func NewInsightProcessorBuilder() *InsightProcessorBuilder {
	return &InsightProcessorBuilder{
		capacity:      DefaultCapacity,
		flushInterval: DefaultFlushInterval,
		retryInterval: DefaultRetryDelay,
		maxRetryTimes: DefaultRetryTimes,
	}
}

// Capacity sets the inbox capacity
func (i *InsightProcessorBuilder) Capacity(capacity int) *InsightProcessorBuilder {
	if capacity <= 0 {
		i.capacity = DefaultCapacity
	} else {
		i.capacity = capacity
	}
	return i
}

// FlushInterval sets the interval of flush message
func (i *InsightProcessorBuilder) FlushInterval(flushInterval time.Duration) *InsightProcessorBuilder {
	if flushInterval <= 0 {
		i.flushInterval = DefaultFlushInterval
	} else {
		i.flushInterval = flushInterval
	}
	return i
}

// RetryInterval sets the time to wait for next retry if the last sending events failed
func (i *InsightProcessorBuilder) RetryInterval(retryInterval time.Duration) *InsightProcessorBuilder {
	if retryInterval <= 0 {
		i.retryInterval = DefaultRetryDelay
	} else {
		i.retryInterval = retryInterval
	}
	return i
}

// MaxRetryTimes sets max retry times for a failed event sending
func (i *InsightProcessorBuilder) MaxRetryTimes(maxRetryTimes int) *InsightProcessorBuilder {
	if maxRetryTimes <= 0 {
		i.maxRetryTimes = DefaultRetryTimes
	} else {
		i.maxRetryTimes = maxRetryTimes
	}
	return i
}

// CreateInsightProcessor creates an instance of interfaces.InsightProcessor
func (i *InsightProcessorBuilder) CreateInsightProcessor(context Context) (InsightProcessor, error) {
	sender, err := i.CreateInsightEventSender(context)
	if err != nil {
		return nil, err
	}
	capacity := i.capacity
	if capacity > DefaultCapacity {
		capacity = DefaultCapacity
	}
	flushInterval := i.flushInterval
	if flushInterval > 3*time.Second {
		flushInterval = 3 * time.Second
	}
	return insight.NewEventProcessor(context, sender, capacity, flushInterval), nil
}

// CreateInsightEventSender creates an instance of interfaces.Sender
func (i *InsightProcessorBuilder) CreateInsightEventSender(context Context) (Sender, error) {
	network := context.GetNetwork()
	client, ok := network.GetHTTPClient().(*http.Client)
	if !ok {
		return nil, fmt.Errorf("non supported HTTP Client")
	}
	headers := network.GetHeaders(nil)
	retryInterval := i.retryInterval
	if retryInterval > time.Second {
		retryInterval = time.Second
	}
	maxRetryTimes := i.maxRetryTimes
	if maxRetryTimes > 3 {
		maxRetryTimes = 3
	}
	return insight.NewEventSenderImp(client, headers, retryInterval, maxRetryTimes), nil
}

type nullInsightProcessorBuilder struct{}

func ExternalEventTrack() InsightProcessorFactory {
	return &nullInsightProcessorBuilder{}
}

func (n *nullInsightProcessorBuilder) CreateInsightProcessor(Context) (InsightProcessor, error) {
	return insight.NewNullEventProcessor(), nil
}
