package interfaces

import "io"

type Event interface {
	IsSendEvent() bool
	Add(ele interface{}) Event
}

type InsightProcessor interface {
	io.Closer
	Send(event Event)
	Flush()
}

type InsightProcessorFactory interface {
	CreateInsightProcessor(context Context) (InsightProcessor, error)
}

type InsightEventSenderFactory interface {
	CreateInsightEventSender(context Context) (Sender, error)
}

type InsightProcessorAndEventSenderFactory interface {
	InsightProcessorFactory
	InsightEventSenderFactory
}
