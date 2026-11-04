package data

import (
	"encoding/json"
	"github.com/featbit/featbit-go-sdk/v2/interfaces"
	"time"
)

type FeatureFlag struct {
	Id                  string       `json:"id"`
	Deleted             bool         `json:"isArchived"`
	Enabled             bool         `json:"isEnabled"`
	InsightsEnabled     bool         `json:"insightsEnabled"`
	Name                string       `json:"name"`
	Key                 string       `json:"key"`
	VariationType       string       `json:"variationType"`
	DisabledVariationId string       `json:"disabledVariationId"`
	Variations          []Variation  `json:"variations"`
	TargetUsers         []TargetUser `json:"targetUsers"`
	Rules               []TargetRule `json:"rules"`
	Fallthrough         Fallthrough  `json:"fallthrough"`
	Tags                []string     `json:"tags"`
	timestamp           int64
	variationMap        map[string]Variation
}

func (f *FeatureFlag) GetId() string {
	return f.Key
}

func (f *FeatureFlag) IsArchived() bool {
	return f.Deleted
}

func (f *FeatureFlag) GetTimestamp() int64 {
	return f.timestamp
}

func (f *FeatureFlag) GetType() int {
	return 100
}

func (f *FeatureFlag) UnmarshalJSON(bytes []byte) error {
	type tmpFeatureFlag FeatureFlag
	tmp := struct {
		*tmpFeatureFlag
		UpdatedAt       time.Time   `json:"updatedAt"`
		InsightsEnabled interface{} `json:"insightsEnabled"`
	}{
		tmpFeatureFlag: (*tmpFeatureFlag)(f),
	}
	if err := json.Unmarshal(bytes, &tmp); err != nil {
		return err
	}
	// Only an explicit boolean false disables collection, including when reusing an object.
	f.InsightsEnabled = true
	if enabled, ok := tmp.InsightsEnabled.(bool); ok {
		f.InsightsEnabled = enabled
	}
	f.timestamp = tmp.UpdatedAt.UnixNano() / int64(time.Millisecond)
	if !f.Deleted {
		f.variationMap = make(map[string]Variation, len(f.Variations))
		for _, variation := range f.Variations {
			f.variationMap[variation.Id] = variation
		}
	}
	return nil
}

func (f *FeatureFlag) ToArchivedItem() interfaces.Item {
	return NewArchivedItem(f.Key, f.timestamp)
}

func (f *FeatureFlag) GetFlagValue(variationId string) string {
	if variation, ok := f.variationMap[variationId]; ok {
		return variation.Value
	}
	return ""
}

func (f *FeatureFlag) ToFeatureFlagMetadata() interfaces.FeatureFlagMetadata {
	return interfaces.FeatureFlagMetadata{
		Id:            f.Id,
		Deleted:       f.Deleted,
		Enabled:       f.Enabled,
		Name:          f.Name,
		Key:           f.Key,
		VariationType: f.VariationType,
		Tags:          f.Tags,
	}
}
