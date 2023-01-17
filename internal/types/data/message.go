package data

const (
	pingMessageType = "ping"
	syncMessageType = "data-sync"
	FullOp          = "full"
	PatchOp         = "patch"
)

type Message struct {
	MessageType string `json:"messageType"`
}

func (m *Message) IsSyncMessage() bool {
	return m.MessageType == syncMessageType
}

type SyncMessage struct {
	Message
	Data TimestampData `json:"data"`
}

func NewPingMessage() *SyncMessage {
	return &SyncMessage{
		Message: Message{MessageType: pingMessageType},
	}
}

func NewSyncMessage(timestamp int64) *SyncMessage {
	return &SyncMessage{
		Message: Message{MessageType: syncMessageType},
		Data:    TimestampData{Timestamp: timestamp},
	}
}

type All struct {
	Message
	Data Data `json:"data"`
}

func (a *All) IsProcessData() bool {
	return a.IsSyncMessage() && (a.Data.EventType == FullOp || a.Data.EventType == PatchOp)
}
