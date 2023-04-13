package datasynchronization

import (
	"encoding/json"
	"github.com/featbit/featbit-go-sdk/fixtures"
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal/types/data"
	"time"
)

type MockStreaming struct {
	success         bool
	loadData        bool
	waitTime        time.Duration
	realDataUpdator DataUpdater
	initialized     bool
}

func (m *MockStreaming) Close() error {
	return nil
}

func (m *MockStreaming) IsInitialized() bool {
	return m.initialized
}

func (m *MockStreaming) Start() <-chan struct{} {
	ret := make(chan struct{})
	go func() {
		time.Sleep(m.waitTime)
		if m.success {
			m.initialized = true
			if m.loadData {
				jsonBytes, _ := fixtures.LoadFBClientTestData()
				var all data.All
				_ = json.Unmarshal(jsonBytes, &all)
				m.realDataUpdator.Init(all.Data.ToStorageType(), all.Data.GetTimestamp())
				m.realDataUpdator.UpdateStatus(OKState())
			}
		}
		close(ret)
	}()
	return ret
}

type MockStreamingBuilder struct {
	success  bool
	loadData bool
	waitTime time.Duration
}

func NewMockStreamingBuilder(success bool, loadDate bool, waitTime time.Duration) *MockStreamingBuilder {
	return &MockStreamingBuilder{success: success, loadData: loadDate, waitTime: waitTime}
}

func (m *MockStreamingBuilder) CreateDataSynchronizer(_ Context, dataUpdater DataUpdater) (DataSynchronizer, error) {
	if m.waitTime <= 0 {
		m.waitTime = 100 * time.Millisecond
	}
	return &MockStreaming{success: m.success, loadData: m.loadData, waitTime: m.waitTime, realDataUpdator: dataUpdater}, nil
}
