package interfaces

import "io"

// Category represents a separated namespace of storable data items.
// The SDK passes instances of this type to the data store to specify whether it is referring to
// a feature flag, a user segment, etc
type Category interface {
	GetName() string
	GetTag() string
}

type Item interface {
	GetId() string
	IsArchived() bool
	GetTimestamp() int64
	GetType() int
	ToArchivedItem() Item
}

// DataStorage Interface for a data storage that holds feature flags, user segments or any other related data received by the SDK.
// Ordinarily, the only implementations of this interface are the default in-memory implementation,
// which holds references to actual SDK data model objects.
// Note that all implementations should permit concurrent access and updates.
type DataStorage interface {
	io.Closer

	// Init Overwrites the storage with a set of items for each collection, if the new version > the old one
	Init(allDate map[Category]map[string]Item, version int64) error

	// Upsert updates or inserts an item in the specified collection. For updates, the object will only be
	// updated if the existing version is less than the new version; for inserts, if the version > the existing one, it will replace
	// the existing one.
	// The SDK may pass an Item that contains an archived object, in that case, assuming the version is greater than any existing version of that item,
	// the store should retain a placeholder rather than simply not storing anything.
	Upsert(category Category, key string, item Item, version int64) (bool, error)

	// Get retrieves an item from the specified collection, if available.
	// If the item has been achieved and the store contains an achieved placeholder, but it should return null
	Get(category Category, key string) (Item, error)

	// GetAll retrieves all items from the specified collection.
	// If the store contains placeholders for deleted items, it should filter them in the results.
	GetAll(category Category) (map[string]Item, error)

	// IsInitialized checks whether this store has been initialized with any data yet.
	IsInitialized() bool

	// GetVersion returns the latest version of storage
	GetVersion() int64
}

// DataStorageFactory Interface for a factory that creates some implementation of DataStorage
type DataStorageFactory interface {
	// CreateDataStorage create an implementation of DataStorage
	CreateDataStorage(Context) (DataStorage, error)
}
