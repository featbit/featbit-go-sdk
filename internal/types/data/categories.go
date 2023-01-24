package data

import "github.com/featbit/featbit-go-sdk/interfaces"

// category type for the feature flag
type features struct{}

// GetName returns the unique namespace identifier for feature flag objects.
func (f features) GetName() string {
	return "featureFlags"
}

// GetTag returns the tag of features category
func (f features) GetTag() string {
	return "ff"
}

// category type for the segment
type segments struct{}

// GetName returns the unique namespace identifier for segment objects.
func (s segments) GetName() string {
	return "segments"
}

// GetTag returns the tag of segments category
func (s segments) GetTag() string {
	return "seg"
}

// category type for the data item only used in test units
type datatests struct{}

// GetName returns the unique namespace identifier for test data item objects.
func (d datatests) GetName() string {
	return "datatests"
}

// GetTag returns the tag of datatests category
func (d datatests) GetTag() string {
	return "test"
}

var (
	// Features global instance of features category
	Features interfaces.Category = features{}
	// Segments global instance of segments category
	Segments interfaces.Category = segments{}
	// Datatests Datetests global instance of datatests category
	Datatests interfaces.Category = datatests{}
	AllCats                       = []interfaces.Category{Features, Segments, Datatests}
)
