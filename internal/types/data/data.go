package data

import (
	"encoding/json"
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"math"
	"sort"
)

type Data struct {
	EventType    string        `json:"eventType"`
	FeatureFlags []FeatureFlag `json:"featureFlags"`
	Segments     []Segment     `json:"segments"`
	timestamp    int64
}

func (d *Data) UnmarshalJSON(bytes []byte) error {
	type tmpData Data
	tmp := struct {
		*tmpData
	}{
		tmpData: (*tmpData)(d),
	}
	if err := json.Unmarshal(bytes, &tmp); err != nil {
		return err
	}
	sort.SliceStable(d.FeatureFlags, func(i, j int) bool {
		return d.FeatureFlags[i].GetTimestamp() < d.FeatureFlags[j].GetTimestamp()
	})
	sort.SliceStable(d.Segments, func(i, j int) bool {
		return d.Segments[i].GetTimestamp() < d.Segments[j].GetTimestamp()
	})
	var timestamp1, timestamp2 int64
	if size := len(d.FeatureFlags); size > 0 {
		timestamp1 = d.FeatureFlags[size-1].GetTimestamp()
	}
	if size := len(d.Segments); size > 0 {
		timestamp2 = d.Segments[size-1].GetTimestamp()
	}
	max := math.Max(float64(timestamp1), float64(timestamp2))
	d.timestamp = int64(max)
	return nil
}

func (d *Data) GetTimestamp() int64 {
	return d.timestamp
}

func (d *Data) ToStorageType() map[Category]map[string]Item {
	put := func(container map[string]Item, item Item) {
		if item.IsArchived() {
			container[item.GetId()] = item.ToArchivedItem()
		} else {
			container[item.GetId()] = item
		}
	}
	flags := make(map[string]Item, len(d.FeatureFlags))
	for _, flag := range d.FeatureFlags {
		put(flags, &flag)
	}
	segments := make(map[string]Item, len(d.Segments))
	for _, segment := range d.Segments {
		put(segments, &segment)
	}
	data := make(map[Category]map[string]Item, 2)
	data[Features] = flags
	data[Segments] = segments
	return data
}

type TimestampData struct {
	Timestamp int64 `json:"timestamp"`
}
