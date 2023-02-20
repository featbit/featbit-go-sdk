package interfaces

import "io"

// Event interface for the analytics events used in FeatBit
type Event interface {
	// IsSendEvent check if the event will be sent to feature flag center
	IsSendEvent() bool
	// Add adds an element in the Event
	Add(ele interface{}) Event
	// GetKey get the unique key of the Event, only internal use
	GetKey() string
}

// InsightProcessor interface for a component to send analytics events.
type InsightProcessor interface {
	io.Closer
	// Send records an event asynchronously.
	Send(event Event)
	// Flush specifies that any buffered events should be sent as soon as possible, rather than waiting for the next flush interval.
	//
	// This method is asynchronous, so events still may not be sent until a later time
	Flush()
}

// InsightProcessorFactory Interface for a factory that creates an implementation of InsightProcessor
type InsightProcessorFactory interface {
	// CreateInsightProcessor creates an implementation of InsightProcessor
	CreateInsightProcessor(context Context) (InsightProcessor, error)
}

// InsightEventSenderFactory Interface for a factory that creates an implementation of Sender
type InsightEventSenderFactory interface {
	// CreateInsightEventSender creates an implementation of Sender
	CreateInsightEventSender(context Context) (Sender, error)
}

// InsightProcessorAndEventSenderFactory
// see InsightProcessorFactory and InsightEventSenderFactory
type InsightProcessorAndEventSenderFactory interface {
	InsightProcessorFactory
	InsightEventSenderFactory
}
