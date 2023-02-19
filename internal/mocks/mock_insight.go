package mocks

import (
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal/insight"
	"sync"
	"time"
)

type SendingInfo struct {
	events []Event
	size   int
}

func (info SendingInfo) Contains(key string) bool {
	for _, event := range info.events {
		if event.GetKey() == key {
			return true
		}
	}
	return false
}

func (info SendingInfo) Size() int {
	return info.size
}

type MockSender struct {
	buffer    chan SendingInfo
	waitGroup *sync.WaitGroup
	wait      chan struct{}
	err       error
	closeErr  error
	mustWait  bool
	lock      sync.RWMutex
	parseJson func([]byte) []Event
}

func NewMockSender() *MockSender {
	return &MockSender{buffer: make(chan SendingInfo, 100), wait: make(chan struct{})}
}

func (m *MockSender) MustWait() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.mustWait = true
}

func (m *MockSender) NoMoreWait() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.mustWait = false
}

func (m *MockSender) SetWaitGroup(group *sync.WaitGroup) {
	m.waitGroup = group
}

func (m *MockSender) SetErr(err error) {
	m.err = err
}

func (m *MockSender) SetCloseErr(err error) {
	m.closeErr = err
}

func (m *MockSender) SetParseJson(f func([]byte) []Event) {
	m.parseJson = f
}

func (m *MockSender) PostJson(_ string, bytes []byte) ([]byte, error) {
	events := m.parseJson(bytes)
	m.buffer <- SendingInfo{events: events, size: len(events)}
	if m.mustWait {
		m.lock.RLock()
		if m.waitGroup != nil {
			m.waitGroup.Done()
		}
		m.lock.RUnlock()
		<-m.wait
	}
	if m.err != nil {
		return nil, m.err
	}
	return nil, nil
}

func (m *MockSender) Close() error {
	if m.buffer != nil {
		close(m.buffer)
	}
	if m.closeErr != nil {
		return m.closeErr
	}
	return nil
}

func (m *MockSender) Completed() {
	if m.wait != nil {
		close(m.wait)
	}
}

func (m *MockSender) GetLatestSendingInfo(timeout time.Duration) (SendingInfo, bool) {
	select {
	case info, ok := <-m.buffer:
		return info, ok
	case <-time.After(timeout):
		return SendingInfo{}, false
	}
}

type MockInsightProcessorFactory struct {
	sender        *MockSender
	capacity      int
	flushInterval time.Duration
}

func NewMockInsightProcessorFactory(sender *MockSender, capacity int, flushInterval time.Duration) *MockInsightProcessorFactory {
	s, c, fi := sender, capacity, flushInterval
	if s == nil {
		s = NewMockSender()
	}
	if c <= 0 {
		c = 100
	}
	if fi <= 0 {
		fi = 100 * time.Millisecond
	}
	return &MockInsightProcessorFactory{sender: s, capacity: c, flushInterval: fi}
}

func (m *MockInsightProcessorFactory) CreateInsightProcessor(ctx Context) (InsightProcessor, error) {
	return insight.NewEventProcessor(ctx, m.sender, m.capacity, m.flushInterval), nil
}
