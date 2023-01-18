package interfaces

type Event interface {
	IsSendEvent() bool
	Add(ele interface{}) Event
}
