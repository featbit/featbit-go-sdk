package data

import (
	"encoding/json"
	"github.com/featbit/featbit-go-sdk/interfaces"
	"time"
)

const (
	SegmentExcludeUser = iota
	SegmentIncludeUser
	SegmentFallthrough
)

type Segment struct {
	Id          string       `json:"id"`
	Deleted     bool         `json:"isArchived"`
	Rules       []TargetRule `json:"rules"`
	timestamp   int64
	includedSet map[string]struct{}
	excludedSet map[string]struct{}
}

func (s *Segment) GetId() string {
	return s.Id
}

func (s *Segment) IsArchived() bool {
	return s.Deleted
}

func (s *Segment) GetTimestamp() int64 {
	return s.timestamp
}

func (s *Segment) GetType() int {
	return 300
}

func (s *Segment) UnmarshalJSON(bytes []byte) error {
	type tmpSegment Segment
	tmp := struct {
		*tmpSegment
		UpdatedAt time.Time `json:"updatedAt"`
		Included  []string  `json:"included"`
		Excluded  []string  `json:"excluded"`
	}{
		tmpSegment: (*tmpSegment)(s),
	}
	if err := json.Unmarshal(bytes, &tmp); err != nil {
		return err
	}
	s.timestamp = tmp.UpdatedAt.UnixNano() / int64(time.Millisecond)
	s.includedSet = make(map[string]struct{}, len(tmp.Included))
	for _, uid := range tmp.Included {
		s.includedSet[uid] = struct{}{}
	}
	s.excludedSet = make(map[string]struct{}, len(tmp.Excluded))
	for _, uid := range tmp.Excluded {
		s.excludedSet[uid] = struct{}{}
	}
	return nil
}

func (s *Segment) ToArchivedItem() interfaces.Item {
	return NewArchivedItem(s.Id, s.timestamp)
}

func (s *Segment) MatchUser(user string) int {
	if _, ok := s.excludedSet[user]; ok {
		return SegmentExcludeUser
	}
	if _, ok := s.includedSet[user]; ok {
		return SegmentIncludeUser
	}
	return SegmentFallthrough
}
