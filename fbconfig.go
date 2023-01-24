package featbit

import (
	"github.com/featbit/featbit-go-sdk/factories"
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"time"
)

type FBConfig struct {
	Offline            bool
	StartWait          time.Duration
	NetworkFactory     NetworkFactory
	DataStorageFactory DataStorageFactory
}

var DefaultFBConfig *FBConfig = &FBConfig{
	Offline:            false,
	StartWait:          15 * time.Second,
	NetworkFactory:     factories.NewNetworkBuilder(),
	DataStorageFactory: factories.NewInMemoryStorageBuilder(),
}
