package data

import (
	"encoding/json"
	"github.com/featbit/featbit-go-sdk/interfaces"
	"time"
)

type FeatureFlag struct {
	Id                    string       `json:"id"`
	Deleted               bool         `json:"isArchived"`
	ExptIncludeAllTargets bool         `json:"exptIncludeAllTargets"`
	Enabled               bool         `json:"isEnabled"`
	Name                  string       `json:"name"`
	Key                   string       `json:"key"`
	VariationType         string       `json:"variationType"`
	DisabledVariationId   string       `json:"disabledVariationId"`
	Variations            []Variation  `json:"variations"`
	TargetUsers           []TargetUser `json:"targetUsers"`
	Rules                 []TargetRule `json:"rules"`
	Fallthrough           Fallthrough  `json:"fallthrough"`
	timestamp             int64
	variationMap          map[string]Variation
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
		UpdatedAt time.Time `json:"updatedAt"`
	}{
		tmpFeatureFlag: (*tmpFeatureFlag)(f),
	}
	if err := json.Unmarshal(bytes, &tmp); err != nil {
		return err
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
