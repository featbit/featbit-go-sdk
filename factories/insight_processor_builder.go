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

type InsightProcessorBuilder struct {
	capacity      int
	flushInterval time.Duration
	retryInterval time.Duration
	maxRetryTimes int
}

func (i *InsightProcessorBuilder) Capacity(capacity int) *InsightProcessorBuilder {
	if capacity <= 0 {
		i.capacity = DefaultCapacity
	} else {
		i.capacity = capacity
	}
	return i
}

func (i *InsightProcessorBuilder) FlushInterval(flushInterval time.Duration) *InsightProcessorBuilder {
	if flushInterval <= 0 {
		i.flushInterval = DefaultFlushInterval
	} else {
		i.flushInterval = flushInterval
	}
	return i
}

func (i *InsightProcessorBuilder) RetryInterval(retryInterval time.Duration) *InsightProcessorBuilder {
	if retryInterval <= 0 {
		i.retryInterval = DefaultRetryDelay
	} else {
		i.retryInterval = retryInterval
	}
	return i
}

func (i *InsightProcessorBuilder) MaxRetryTimes(maxRetryTimes int) *InsightProcessorBuilder {
	if maxRetryTimes <= 0 {
		i.maxRetryTimes = DefaultRetryTimes
	} else {
		i.maxRetryTimes = maxRetryTimes
	}
	return i
}

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
