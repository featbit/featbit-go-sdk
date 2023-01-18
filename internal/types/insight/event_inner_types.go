package insight

import "time"

type UserAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type EventUser struct {
	KeyId string          `json:"keyId"`
	Name  string          `json:"name"`
	Attrs []UserAttribute `json:"customizedProperties"`
}

func (u EventUser) isValid() bool {
	return !(u.KeyId == "" || u.Name == "")
}

type EventVariation struct {
	Id     string `json:"id"`
	Value  string `json:"value"`
	Reason string `json:"reason"`
}

type EventFlag struct {
	FlagKey          string         `json:"featureFlagKey"`
	SendToExperiment bool           `json:"sendToExperiment"`
	Timestamp        int64          `json:"timestamp"`
	Variation        EventVariation `json:"variation"`
}

func NewEventFlag(flagKey string, sendToExpt bool, variationId string, variation string, reason string) EventFlag {
	return EventFlag{
		Variation: EventVariation{
			Id:     variationId,
			Value:  variation,
			Reason: reason,
		},
		FlagKey:          flagKey,
		SendToExperiment: sendToExpt,
		Timestamp:        time.Now().UnixNano() / int64(time.Millisecond),
	}
}

type Metric struct {
	Route        string  `json:"route"`
	Type         string  `json:"type"`
	EventName    string  `json:"eventName"`
	NumericValue float64 `json:"numericValue"`
	AppType      string  `json:"appType"`
	Timestamp    int64   `json:"timestamp"`
}

func NewMetric(evt string, weight float64) Metric {
	return Metric{
		Route:        "index/metric",
		Type:         "CustomEvent",
		EventName:    evt,
		NumericValue: weight,
		AppType:      "goserverside",
		Timestamp:    time.Now().UnixNano() / int64(time.Millisecond),
	}
}
