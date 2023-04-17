package datasynchronization

import (
	. "github.com/featbit/featbit-go-sdk/interfaces"
)

type NullDataSynchronizer struct {
	realDataUpdater DataUpdater
}

func NewNullDataSynchronizer(dataUpdater DataUpdater) *NullDataSynchronizer {
	return &NullDataSynchronizer{realDataUpdater: dataUpdater}
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
	n.realDataUpdater.UpdateStatus(OKState())
	return ready
}
