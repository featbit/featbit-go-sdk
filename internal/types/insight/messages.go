package insight

import . "github.com/featbit/featbit-go-sdk/interfaces"

type EventMessage interface{}

type SendingMessage struct {
	event Event
}

func (s SendingMessage) GetEvent() Event {
	return s.event
}

func NewSendingEvent(event Event) SendingMessage {
	return SendingMessage{event: event}
}

type FlushingMessage struct{}

type ShutdownMessage struct {
	waitCh chan struct{}
}

func (s ShutdownMessage) Completed() {
	s.waitCh <- struct{}{}
	close(s.waitCh)
}

func (s ShutdownMessage) GetWaitCh() <-chan struct{} {
	return s.waitCh
}

func NewShutdownMessage() ShutdownMessage {
	return ShutdownMessage{make(chan struct{})}
}
