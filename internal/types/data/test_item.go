package data

import (
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/google/uuid"
	"time"
)

type TestItem struct {
	deleted   bool
	timestamp int64
	id        string
}

func NewTestItem(deleted bool) *TestItem {
	return &TestItem{deleted: deleted,
		timestamp: time.Now().UnixNano(),
		id:        uuid.New().String(),
	}
}

func (t *TestItem) GetId() string {
	return t.id
}

func (t *TestItem) IsArchived() bool {
	return t.deleted
}

func (t *TestItem) GetTimestamp() int64 {
	return t.timestamp
}

func (t *TestItem) GetType() int {
	return 400
}

func (t *TestItem) ToArchivedItem() Item {
	return NewArchivedItem(t.id, t.timestamp)
}
