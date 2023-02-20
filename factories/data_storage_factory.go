package factories

import (
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal/datastorage"
)

type InMemoryStorageBuilder struct{}

func NewInMemoryStorageBuilder() InMemoryStorageBuilder {
	return InMemoryStorageBuilder{}
}

func (i InMemoryStorageBuilder) CreateDataStorage(Context) (DataStorage, error) {
	return datastorage.NewInMemoryDataStorage(), nil
}
