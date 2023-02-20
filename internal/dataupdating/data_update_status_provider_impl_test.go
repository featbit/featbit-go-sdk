package dataupdating

import (
	"github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal/datastorage"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestWaitFor(t *testing.T) {
	t.Run("already OK state", func(t *testing.T) {
		dataStorage := datastorage.NewMockDataStorage(datastorage.NewInMemoryDataStorage())
		dataUpdater := NewDataUpdaterImpl(dataStorage)
		dataUpdateStatusProvider := NewDataUpdateStatusProviderImpl(dataUpdater)

		dataUpdater.UpdateStatus(interfaces.OKState())
		assert.True(t, dataUpdateStatusProvider.WaitFor(interfaces.OK, 100*time.Millisecond))
	})
	t.Run("wait for ok", func(t *testing.T) {
		dataStorage := datastorage.NewMockDataStorage(datastorage.NewInMemoryDataStorage())
		dataUpdater := NewDataUpdaterImpl(dataStorage)
		dataUpdateStatusProvider := NewDataUpdateStatusProviderImpl(dataUpdater)
		go func() {
			time.Sleep(50 * time.Millisecond)
			dataUpdater.UpdateStatus(interfaces.OKState())
		}()
		t1 := time.Now()
		assert.True(t, dataUpdateStatusProvider.WaitFor(interfaces.OK, 100*time.Millisecond))
		t2 := time.Now()
		duration := t2.Sub(t1)
		assert.True(t, duration >= 50*time.Millisecond)
	})
	t.Run("wait for ok but time out", func(t *testing.T) {
		dataStorage := datastorage.NewMockDataStorage(datastorage.NewInMemoryDataStorage())
		dataUpdater := NewDataUpdaterImpl(dataStorage)
		dataUpdateStatusProvider := NewDataUpdateStatusProviderImpl(dataUpdater)
		assert.False(t, dataUpdateStatusProvider.WaitFor(interfaces.OK, 10*time.Millisecond))
	})
	t.Run("wait for ok but off comes", func(t *testing.T) {
		dataStorage := datastorage.NewMockDataStorage(datastorage.NewInMemoryDataStorage())
		dataUpdater := NewDataUpdaterImpl(dataStorage)
		dataUpdateStatusProvider := NewDataUpdateStatusProviderImpl(dataUpdater)
		go func() {
			time.Sleep(50 * time.Millisecond)
			dataUpdater.UpdateStatus(interfaces.NormalOFFState())
		}()
		t1 := time.Now()
		assert.False(t, dataUpdateStatusProvider.WaitFor(interfaces.OK, 100*time.Millisecond))
		t2 := time.Now()
		duration := t2.Sub(t1)
		assert.True(t, duration >= 50*time.Millisecond)
	})
}
