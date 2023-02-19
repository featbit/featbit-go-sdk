package insight

import (
	"encoding/json"
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal/types/insight"
	"github.com/featbit/featbit-go-sdk/internal/util/log"
	"strconv"
	"sync"
	"time"
)

const (
	MaxEventSizePerRequest = 50
	MaxFlushWorkersNumber  = 5
)

type nextFlushBuffer struct {
	events []Event
}

func newNextFlushBuffer(capacity int) *nextFlushBuffer {
	return &nextFlushBuffer{events: make([]Event, 0, capacity)}
}

func (b *nextFlushBuffer) isEmpty() bool {
	return len(b.events) == 0
}

func (b *nextFlushBuffer) add(event Event) {
	b.events = append(b.events, event)
}

func (b *nextFlushBuffer) getPayload() *payload {
	var copied []Event
	if len(b.events) > 0 {
		copied = make([]Event, len(b.events))
		copy(copied, b.events)
	}
	return &payload{events: copied}
}

func (b *nextFlushBuffer) clear() {
	for i := range b.events {
		b.events[i] = nil
	}
	b.events = b.events[:0]
}

type payload struct {
	events []Event
}

func (p *payload) split(size int) [][]Event {
	if len(p.events) == 0 {
		return nil
	}
	var n int = size
	if n <= 0 {
		n = MaxEventSizePerRequest
	}
	parts := len(p.events) / n
	r := len(p.events) % n
	if r > 0 {
		parts += 1
	}
	ret := make([][]Event, parts)
	for i := 0; i < parts; i++ {
		start := i * n
		if i == parts-1 {
			ret[i] = p.events[start:]
		} else {
			ret[i] = p.events[start : (i+1)*n]
		}
	}
	return ret
}

type eventDispatcher struct {
	buffer   *nextFlushBuffer
	outboxCh chan *payload
	permits  *sync.WaitGroup
	closed   bool
}

// blocks until a message is available and then:
// 1: transfer the events to event buffer
// 2: try to flush events to feature flag center if a flush message arrives
// 3: wait for releasing resources if a shutdown arrives
func (ed *eventDispatcher) runDispatchEvents(inboxCh <-chan insight.EventMessage, flushInterval time.Duration) {
	if ed.closed {
		return
	}
	log.LogDebug("event dispatcher is working")
	log.LogDebug("flush ticker is starting")
	flushScheduler := time.NewTicker(flushInterval)
	for {
		select {
		case message := <-inboxCh:
			switch msg := message.(type) {
			case insight.SendingMessage:
				ed.putEventToNextBuffer(msg.GetEvent())
			case insight.FlushingMessage:
				ed.triggerFlush()
			case insight.ShutdownMessage:
				log.LogDebug("event dispatcher is stopping")
				ed.closed = true
				log.LogDebug("flush ticker is over")
				flushScheduler.Stop()
				ed.permits.Wait()
				close(ed.outboxCh)
				msg.Completed()
				return
			}
		case <-flushScheduler.C:
			ed.triggerFlush()
		}
	}
}

func (ed *eventDispatcher) triggerFlush() {
	if ed.closed || ed.buffer.isEmpty() {
		return
	}
	//get all the current events from event buffer
	payload := ed.buffer.getPayload()
	// increment the count of active flushe runner
	ed.permits.Add(1)
	select {
	case ed.outboxCh <- payload:
		log.LogDebug("trigger flush")
		// clear unused buffer for next flush
		ed.buffer.clear()
	default:
		// if no more available flush workers, the buffer will be merged in the next flush
		// we can't start a flush right now because we're waiting for one of flush worker to pick up the last one
		// decrease the delta that we incremented just now
		ed.permits.Done()
	}
}

func (ed *eventDispatcher) putEventToNextBuffer(event Event) {
	if ed.closed {
		return
	}
	if event.IsSendEvent() {
		log.LogDebug("put event to buffer")
		ed.buffer.add(event)
	}
}

func runFlashRunner(name string, eventUri string, sender Sender, outboxCh <-chan *payload, permits *sync.WaitGroup) {
	log.LogDebug("%s is starting", name)
	for {
		payloads, running := <-outboxCh
		if !running {
			// outbox closed - we're shutting down
			log.LogDebug("%s is over", name)
			return
		}
		// split the payload into small partitions and send them to feature flag center
		for _, payload := range payloads.split(MaxEventSizePerRequest) {
			log.LogDebug("payload size: %v", len(payload))
			jsonBytes, _ := json.Marshal(payload)
			_, _ = sender.PostJson(eventUri, jsonBytes)
		}
		permits.Done()
	}
}

func startEventDispatcher(context Context, inboxCh <-chan insight.EventMessage, sender Sender, capacity int, flushInterval time.Duration) {
	ed := &eventDispatcher{
		buffer:   newNextFlushBuffer(capacity),
		outboxCh: make(chan *payload),
		permits:  &sync.WaitGroup{},
	}
	for i := 0; i < MaxFlushWorkersNumber; i++ {
		name := "flush-worker-" + strconv.Itoa(i)
		go runFlashRunner(name, context.GetEventUri(), sender, ed.outboxCh, ed.permits)
	}
	go ed.runDispatchEvents(inboxCh, flushInterval)

}

type EventProcessor struct {
	inboxCh chan insight.EventMessage
	// close processor action should call only one time
	closeOnce sync.Once
	// processor closed sig
	processorClosed bool
	sender          Sender
}

func NewEventProcessor(context Context, sender Sender, capacity int, flushInterval time.Duration) *EventProcessor {
	inboxCh := make(chan insight.EventMessage, capacity)
	startEventDispatcher(context, inboxCh, sender, capacity, flushInterval)
	return &EventProcessor{inboxCh: inboxCh, sender: sender}
}

func (ep *EventProcessor) putMsgToBox(msg insight.EventMessage) bool {
	for {
		select {
		case ep.inboxCh <- msg:
			return true
		default:
			if _, ok := msg.(insight.ShutdownMessage); ok {
				continue
			}
			// if it reaches here, it means the application is probably doing tons of flag evaluations across many threads.
			// So if we wait for a space in the inbox, we risk a very serious slowdown of the app.
			// To avoid that, we'll just drop the event or you can increase the capacity of inbox
			log.LogWarn("FB JAVA SDK: events are being produced faster than they can be processed; some events will be dropped")
			return false
		}
	}
}

func (ep *EventProcessor) Close() error {
	ep.closeOnce.Do(func() {
		log.LogInfo("FB GO SDK: insight processor is stopping")
		ep.processorClosed = true
		//flush all the left events
		ep.putMsgToBox(insight.FlushingMessage{})
		//shutdown, clear all the threads
		shutdown := insight.NewShutdownMessage()
		ep.putMsgToBox(shutdown)
		<-shutdown.GetWaitCh()
		_ = ep.sender.Close()
	})
	return nil
}

func (ep *EventProcessor) Send(event Event) {
	if ep.processorClosed || event == nil {
		return
	}
	switch event.(type) {
	case *insight.UserEvent, *insight.FlagEvent, *insight.MetricEvent:
		ep.putMsgToBox(insight.NewSendingEvent(event))
	default:
		log.LogWarn("ignore event")
	}
}

func (ep *EventProcessor) Flush() {
	if ep.processorClosed {
		return
	}
	ep.putMsgToBox(insight.FlushingMessage{})
}
