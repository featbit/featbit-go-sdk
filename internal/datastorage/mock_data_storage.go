package datastorage

import (
	"encoding/json"
	"fmt"
	"github.com/featbit/featbit-go-sdk/fixtures"
	"github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal/types/data"
	"sync"
)

type MockDataStorage struct {
	realDataStorage interfaces.DataStorage
	fakeErr         error
	lock            sync.Mutex
}

func NewMockDataStorage(realDataStorage interfaces.DataStorage) *MockDataStorage {
	return &MockDataStorage{realDataStorage: realDataStorage}
}

func (m *MockDataStorage) LoadData() error {
	if m.realDataStorage != nil {
		jsonBytes, _ := fixtures.LoadFBClientTestData()
		var all data.All
		_ = json.Unmarshal(jsonBytes, &all)
		return m.realDataStorage.Init(all.Data.ToStorageType(), all.Data.GetTimestamp())

	}
	return fmt.Errorf("nil data storage")
}

func (m *MockDataStorage) Close() error {
	return nil
}

func (m *MockDataStorage) Init(allData map[interfaces.Category]map[string]interfaces.Item, version int64) error {
	_ = m.realDataStorage.Init(allData, version)
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.fakeErr
}

func (m *MockDataStorage) Upsert(category interfaces.Category, key string, item interfaces.Item, version int64) (bool, error) {
	_, _ = m.realDataStorage.Upsert(category, key, item, version)
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.fakeErr == nil, m.fakeErr
}

func (m *MockDataStorage) Get(category interfaces.Category, key string) (interfaces.Item, error) {
	item, err := m.realDataStorage.Get(category, key)
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.fakeErr == nil {
		return item, err
	}
	return nil, m.fakeErr
}

func (m *MockDataStorage) GetAll(category interfaces.Category) (map[string]interfaces.Item, error) {
	items, err := m.realDataStorage.GetAll(category)
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.fakeErr == nil {
		return items, err
	}
	return nil, m.fakeErr
}

func (m *MockDataStorage) IsInitialized() bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.fakeErr == nil && m.realDataStorage.IsInitialized()
}

func (m *MockDataStorage) GetVersion() int64 {
	return m.realDataStorage.GetVersion()
}

func (m *MockDataStorage) SetErr(err error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.fakeErr = err
}

type MockDataStorageBuilder struct {
	fakeErr error
}

func NewMockDataStorageBuilder() *MockDataStorageBuilder {
	return &MockDataStorageBuilder{}
}

func (m *MockDataStorageBuilder) CreateDataStorage(_ interfaces.Context) (interfaces.DataStorage, error) {
	if m.fakeErr != nil {
		return &MockDataStorage{realDataStorage: NewInMemoryDataStorage(), fakeErr: m.fakeErr}, nil
	}
	return &MockDataStorage{realDataStorage: NewInMemoryDataStorage()}, nil
}
