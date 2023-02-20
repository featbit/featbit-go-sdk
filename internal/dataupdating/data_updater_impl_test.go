package dataupdating

import (
	"fmt"
	"github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal/datastorage"
	"github.com/featbit/featbit-go-sdk/internal/types/data"
	"github.com/stretchr/testify/assert"
	"testing"
)

var item1 = data.NewTestItem(false)

func TestInit(t *testing.T) {
	// data
	items := map[string]interfaces.Item{item1.GetId(): item1}
	all := map[interfaces.Category]map[string]interfaces.Item{data.Datatests: items}
	t.Run("init", func(t *testing.T) {
		dataStorage := datastorage.NewMockDataStorage(datastorage.NewInMemoryDataStorage())
		dataUpdater := NewDataUpdaterImpl(dataStorage)
		ok := dataUpdater.Init(all, int64(1))
		if ok {
			dataUpdater.UpdateStatus(interfaces.OKState())
		}
		assert.True(t, ok)
		assert.True(t, dataUpdater.StorageInitialized())
		assert.Equal(t, interfaces.OK, dataUpdater.currentState.StateType)
		assert.Equal(t, int64(1), dataUpdater.GetVersion())
		item, _ := dataStorage.Get(data.Datatests, item1.GetId())
		assert.Equal(t, item1, item)
	})
	t.Run("initWithError", func(t *testing.T) {
		mockDataStorage := datastorage.NewMockDataStorage(datastorage.NewInMemoryDataStorage())
		mockDataStorage.SetErr(fmt.Errorf("fake error"))
		dataUpdater := NewDataUpdaterImpl(mockDataStorage)
		ok := dataUpdater.Init(all, int64(1))
		assert.False(t, ok)
		assert.False(t, dataUpdater.StorageInitialized())
		assert.Equal(t, interfaces.INITIALIZING, dataUpdater.getCurrentState().StateType)
		assert.Equal(t, interfaces.DataStorageInitError, dataUpdater.getCurrentState().ErrorTrack.ErrorType)
	})
}

func TestUpsert(t *testing.T) {
	t.Run("upsert", func(t *testing.T) {
		dataStorage := datastorage.NewMockDataStorage(datastorage.NewInMemoryDataStorage())
		dataUpdater := NewDataUpdaterImpl(dataStorage)
		ok := dataUpdater.Upsert(data.Datatests, item1.GetId(), item1, int64(1))
		if ok {
			dataUpdater.UpdateStatus(interfaces.OKState())
		}
		assert.True(t, ok)
		assert.True(t, dataUpdater.StorageInitialized())
		assert.Equal(t, interfaces.OK, dataUpdater.currentState.StateType)
		assert.Equal(t, int64(1), dataUpdater.GetVersion())
		item, _ := dataStorage.Get(data.Datatests, item1.GetId())
		assert.Equal(t, item1, item)
	})
	t.Run("upsertWithError", func(t *testing.T) {
		mockDataStorage := datastorage.NewMockDataStorage(datastorage.NewInMemoryDataStorage())
		mockDataStorage.SetErr(fmt.Errorf("fake error"))
		dataUpdater := NewDataUpdaterImpl(mockDataStorage)
		ok := dataUpdater.Upsert(data.Datatests, item1.GetId(), item1, int64(1))
		assert.False(t, ok)
		assert.False(t, dataUpdater.StorageInitialized())
		assert.Equal(t, interfaces.INITIALIZING, dataUpdater.getCurrentState().StateType)
		assert.Equal(t, interfaces.DataStorageUpdateError, dataUpdater.getCurrentState().ErrorTrack.ErrorType)
	})
}
