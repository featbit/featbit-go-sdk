package factories

import (
	. "github.com/featbit/featbit-go-sdk/v2/interfaces"
	"github.com/featbit/featbit-go-sdk/v2/internal/datastorage"
)

type InMemoryStorageBuilder struct{}

func NewInMemoryStorageBuilder() InMemoryStorageBuilder {
	return InMemoryStorageBuilder{}
}

func (i InMemoryStorageBuilder) CreateDataStorage(Context) (DataStorage, error) {
	return datastorage.NewInMemoryDataStorage(), nil
}
