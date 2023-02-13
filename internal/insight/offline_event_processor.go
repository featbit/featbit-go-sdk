package insight

import (
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"sync"
)

type NullEventProcessor struct{}

var instance *NullEventProcessor
var once sync.Once

func NewNullEventProcessor() *NullEventProcessor {
	once.Do(func() {
		instance = &NullEventProcessor{}
	})
	return instance
}

func (n *NullEventProcessor) Close() error {
	return nil
}

func (n *NullEventProcessor) Send(Event) {}

func (n *NullEventProcessor) Flush() {}
