package interfaces

import "io"

type DataSynchronizer interface {
	io.Closer

	isInitialized() bool

	start(ready chan<- struct{})
}
