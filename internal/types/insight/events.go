package insight

import "github.com/featbit/featbit-go-sdk/interfaces"

type BaseEvent struct {
	User EventUser `json:"user"`
}

type UserEvent struct {
	BaseEvent
}

func (u *UserEvent) IsSendEvent() bool {
	return u.User.isValid()
}

func (u *UserEvent) Add(interface{}) interfaces.Event {
	return u
}

func NewUserEvent(user EventUser) *UserEvent {
	return &UserEvent{
		BaseEvent{User: user},
	}
}

type FlagEvent struct {
	BaseEvent
	Variations []EventFlag `json:"variations"`
}

func (f *FlagEvent) IsSendEvent() bool {
	return f.User.isValid() && len(f.Variations) > 0
}

func (f *FlagEvent) Add(ele interface{}) interfaces.Event {
	flag, ok := ele.(EventFlag)
	if ok {
		f.Variations = append(f.Variations, flag)
	}
	return f
}

func NewFlagEvent(user EventUser) *FlagEvent {
	return &FlagEvent{
		BaseEvent: BaseEvent{User: user},
	}
}

type MetricEvent struct {
	BaseEvent
	Metrics []Metric `json:"metrics"`
}

func (m *MetricEvent) IsSendEvent() bool {
	return m.User.isValid() && len(m.Metrics) > 0
}

func (m *MetricEvent) Add(ele interface{}) interfaces.Event {
	metric, ok := ele.(Metric)
	if ok {
		m.Metrics = append(m.Metrics, metric)
	}
	return m
}

func NewMetricEvent(user EventUser) *MetricEvent {
	return &MetricEvent{
		BaseEvent: BaseEvent{User: user},
	}
}
