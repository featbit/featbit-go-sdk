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
	waitTime        time.Duration
	realDataUpdator DataUpdater
}

func (m *MockStreaming) Close() error {
	return nil
}

func (m *MockStreaming) IsInitialized() bool {
	return m.realDataUpdator.StorageInitialized()
}

func (m *MockStreaming) Start() <-chan struct{} {
	ret := make(chan struct{})
	go func() {
		time.Sleep(m.waitTime)
		if m.success {
			jsonBytes, _ := fixtures.LoadFBClientTestData()
			var all data.All
			_ = json.Unmarshal(jsonBytes, &all)
			m.realDataUpdator.Init(all.Data.ToStorageType(), all.Data.GetTimestamp())
			m.realDataUpdator.UpdateStatus(OKState())
		}
		close(ret)
	}()
	return ret
}

type MockStreamingBuilder struct {
	success  bool
	waitTime time.Duration
}

func NewMockStreamingBuilder(success bool, waitTime time.Duration) *MockStreamingBuilder {
	return &MockStreamingBuilder{success: success, waitTime: waitTime}
}

func (m *MockStreamingBuilder) CreateDataSynchronizer(_ Context, dataUpdater DataUpdater) (DataSynchronizer, error) {
	if m.waitTime <= 0 {
		m.waitTime = 100 * time.Millisecond
	}
	return &MockStreaming{success: true, waitTime: m.waitTime, realDataUpdator: dataUpdater}, nil
}
