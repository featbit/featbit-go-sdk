package interfaces

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
