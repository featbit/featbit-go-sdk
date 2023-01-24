package datastorage

import (
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"sync"
)

type InMemoryDataStorage struct {
	version     int64
	initialized bool
	allData     map[Category]map[string]Item
	lock        sync.RWMutex
}

func (i *InMemoryDataStorage) Close() error {
	return nil
}

func (i *InMemoryDataStorage) Init(allData map[Category]map[string]Item, version int64) error {
	if version <= i.version || len(allData) == 0 {
		return nil
	}
	i.lock.Lock()
	defer i.lock.Unlock()
	i.allData = make(map[Category]map[string]Item, len(allData))
	for cat, items := range allData {
		_items := make(map[string]Item, len(items))
		for key, value := range items {
			_items[key] = value
		}
		i.allData[cat] = _items
	}
	i.initialized = true
	i.version = version
	return nil
}

func (i *InMemoryDataStorage) Upsert(category Category, key string, item Item, version int64) (bool, error) {
	if i.version >= version || item == nil || category == nil || key == "" {
		return false, nil
	}
	i.lock.Lock()
	defer i.lock.Unlock()
	if len(i.allData) == 0 {
		i.allData = make(map[Category]map[string]Item, 1)
	}
	if items, ok := i.allData[category]; ok {
		if oldItem, ok1 := items[key]; ok1 {
			if oldItem.GetTimestamp() < version {
				items[key] = item
			}
		} else {
			items[key] = item
		}
	} else {
		i.allData[category] = map[string]Item{key: item}
	}
	if !i.initialized {
		i.initialized = true
	}
	return true, nil
}

func (i *InMemoryDataStorage) Get(category Category, key string) (Item, error) {
	i.lock.RLock()
	defer i.lock.RUnlock()
	var items map[string]Item
	var res Item
	var ok bool
	if items, ok = i.allData[category]; ok {
		res, ok = items[key]
	}
	if ok && res.IsArchived() {
		res = nil
	}
	return res, nil
}

func (i *InMemoryDataStorage) GetAll(category Category) (map[string]Item, error) {
	i.lock.RLock()
	defer i.lock.RUnlock()
	var res, items map[string]Item
	var ok bool
	if items, ok = i.allData[category]; ok {
		res = make(map[string]Item, len(items))
		for k, v := range items {
			if !v.IsArchived() {
				res[k] = v
			}
		}
	}
	return res, nil
}

func (i *InMemoryDataStorage) IsInitialized() bool {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return i.initialized
}

func (i *InMemoryDataStorage) GetVersion() int64 {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return i.version
}

func NewInMemoryDataStorage() *InMemoryDataStorage {
	return &InMemoryDataStorage{}
}
