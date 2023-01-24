package data

import "github.com/featbit/featbit-go-sdk/interfaces"

type ArchivedItem struct {
	id        string
	timestamp int64
}

func (a *ArchivedItem) ToArchivedItem() interfaces.Item {
	return a
}

func (a *ArchivedItem) GetId() string {
	return a.id
}

func (a *ArchivedItem) IsArchived() bool {
	return true
}

func (a *ArchivedItem) GetTimestamp() int64 {
	return a.timestamp
}

func (a *ArchivedItem) GetType() int {
	return 200
}

func NewArchivedItem(id string, timestamp int64) *ArchivedItem {
	return &ArchivedItem{
		id:        id,
		timestamp: timestamp,
	}
}

type Variation struct {
	Id    string `json:"id"`
	Value string `json:"value"`
}

type TargetUser struct {
	KeyIds      []string `json:"keyIds"`
	VariationId string   `json:"variationId"`
}

type TargetRule struct {
	IncludedInExpt bool               `json:"includedInExpt"`
	Conditions     []Condition        `json:"conditions"`
	Variations     []RolloutVariation `json:"variations"`
	DispatchKey    string             `json:"dispatchKey"`
}

type Condition struct {
	Property string `json:"property"`
	Op       string `json:"op"`
	Value    string `json:"value"`
}

type RolloutVariation struct {
	Id          string    `json:"id"`
	Rollout     []float64 `json:"rollout"`
	ExptRollout float64   `json:"exptRollout"`
}

type Fallthrough struct {
	IncludedInExpt bool               `json:"includedInExpt"`
	Variations     []RolloutVariation `json:"variations"`
	DispatchKey    string             `json:"dispatchKey"`
}
