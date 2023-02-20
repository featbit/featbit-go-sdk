package datastorage

import (
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal/types/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var item1 = data.NewTestItem(false)
var item2 = data.NewTestItem(false)
var item3 = data.NewTestItem(true)

func TestInit(t *testing.T) {
	t.Run("default version", func(t *testing.T) {
		dataStorage := NewInMemoryDataStorage()
		assert.Equal(t, int64(0), dataStorage.GetVersion())
		assert.False(t, dataStorage.initialized)
	})
	t.Run("init", func(t *testing.T) {
		items := map[string]Item{item1.GetId(): item1, item3.GetId(): item3}
		allData := map[Category]map[string]Item{data.Datatests: items}
		dataStorage := NewInMemoryDataStorage()
		require.NoError(t, dataStorage.Init(allData, int64(1)))
		assert.True(t, dataStorage.IsInitialized())
		assert.Equal(t, int64(1), dataStorage.GetVersion())
		item, _ := dataStorage.Get(data.Datatests, item1.GetId())
		assert.Equal(t, item1, item)
		item, _ = dataStorage.Get(data.Datatests, item3.GetId())
		assert.Nil(t, item)
		allItems, _ := dataStorage.GetAll(data.Datatests)
		assert.Equal(t, 1, len(allItems))
	})
	t.Run("invalid init", func(t *testing.T) {
		dataStorage := NewInMemoryDataStorage()
		require.NoError(t, dataStorage.Init(nil, int64(1)))
		assert.False(t, dataStorage.IsInitialized())
		assert.Equal(t, int64(0), dataStorage.GetVersion())
		allData := make(map[Category]map[string]Item, 0)
		require.NoError(t, dataStorage.Init(allData, int64(1)))
		assert.False(t, dataStorage.IsInitialized())
		assert.Equal(t, int64(0), dataStorage.GetVersion())
		items := map[string]Item{item1.GetId(): item1}
		allData = map[Category]map[string]Item{data.Datatests: items}
		require.NoError(t, dataStorage.Init(allData, int64(-1)))
		assert.False(t, dataStorage.IsInitialized())
		assert.Equal(t, int64(0), dataStorage.GetVersion())
		require.NoError(t, dataStorage.Init(allData, int64(1)))
		assert.True(t, dataStorage.IsInitialized())
		assert.Equal(t, int64(1), dataStorage.GetVersion())
		items = map[string]Item{item1.GetId(): item1, item2.GetId(): item2}
		allData = map[Category]map[string]Item{data.Datatests: items}
		require.NoError(t, dataStorage.Init(allData, int64(1)))
		assert.Equal(t, int64(1), dataStorage.GetVersion())
		allItems, _ := dataStorage.GetAll(data.Datatests)
		assert.Equal(t, 1, len(allItems))
	})

}

func TestUpsert(t *testing.T) {
	t.Run("upsert", func(t *testing.T) {
		dataStorage := NewInMemoryDataStorage()
		ok, err := dataStorage.Upsert(data.Datatests, item1.GetId(), item1, int64(1))
		assert.True(t, ok)
		require.NoError(t, err)
		assert.True(t, dataStorage.IsInitialized())
		assert.Equal(t, int64(1), dataStorage.GetVersion())
		item, _ := dataStorage.Get(data.Datatests, item1.GetId())
		assert.Equal(t, item1, item)
		ok, err = dataStorage.Upsert(data.Datatests, item2.GetId(), item2, int64(2))
		assert.True(t, ok)
		require.NoError(t, err)
		assert.Equal(t, int64(2), dataStorage.GetVersion())
		item, _ = dataStorage.Get(data.Datatests, item2.GetId())
		assert.Equal(t, item2, item)
		newItem := data.NewTestItem(false)
		ok, err = dataStorage.Upsert(data.Datatests, item1.GetId(), newItem, int64(3))
		assert.True(t, ok)
		require.NoError(t, err)
		assert.Equal(t, int64(3), dataStorage.GetVersion())
		item, _ = dataStorage.Get(data.Datatests, item1.GetId())
		assert.Equal(t, newItem, item)
	})
	t.Run("invalid upsert", func(t *testing.T) {
		dataStorage := NewInMemoryDataStorage()
		ok, err := dataStorage.Upsert(nil, item1.GetId(), item1, int64(1))
		assert.False(t, ok)
		require.NoError(t, err)
		assert.False(t, dataStorage.IsInitialized())
		assert.Equal(t, int64(0), dataStorage.GetVersion())
		ok, err = dataStorage.Upsert(data.Datatests, "", item1, int64(1))
		assert.False(t, ok)
		require.NoError(t, err)
		assert.False(t, dataStorage.IsInitialized())
		assert.Equal(t, int64(0), dataStorage.GetVersion())
		ok, err = dataStorage.Upsert(data.Datatests, item1.GetId(), nil, int64(1))
		assert.False(t, ok)
		require.NoError(t, err)
		assert.False(t, dataStorage.IsInitialized())
		assert.Equal(t, int64(0), dataStorage.GetVersion())
		ok, err = dataStorage.Upsert(data.Datatests, item1.GetId(), item1, int64(-1))
		assert.False(t, ok)
		require.NoError(t, err)
		assert.False(t, dataStorage.IsInitialized())
		assert.Equal(t, int64(0), dataStorage.GetVersion())

	})
}
