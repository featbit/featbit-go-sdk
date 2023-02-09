package interfaces

import "io"

// DataSynchronizer Interface to receive updates to feature flags, user segments, and anything
// else that might come from feature flag center, and passes them to a DataStorage
type DataSynchronizer interface {
	io.Closer

	// IsInitialized returns true once the client has been initialized and will never return false again.
	IsInitialized() bool

	// Start starts the client update processing.
	Start() <-chan struct{}
}

// DataSynchronizerFactory Interface for a factory that creates some implementation of DataSynchronizer
type DataSynchronizerFactory interface {
	// CreateDataSynchronizer creates an implementation instance.
	CreateDataSynchronizer(context Context, dataUpdater DataUpdater) (DataSynchronizer, error)
}
