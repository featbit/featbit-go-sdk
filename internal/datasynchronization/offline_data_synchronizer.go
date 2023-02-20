package datasynchronization

import "sync"

type NullDataSynchronizer struct{}

var instance *NullDataSynchronizer
var once sync.Once

func NewNullDataSynchronizer() *NullDataSynchronizer {
	once.Do(func() {
		instance = &NullDataSynchronizer{}
	})
	return instance
}

func (n *NullDataSynchronizer) Close() error {
	return nil
}

func (n *NullDataSynchronizer) IsInitialized() bool {
	return true
}

func (n *NullDataSynchronizer) Start() <-chan struct{} {
	ready := make(chan struct{})
	close(ready)
	return ready
}
