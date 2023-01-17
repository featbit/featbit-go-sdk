package data

import "github.com/featbit/featbit-go-sdk/interfaces"

type features struct{}

func (f features) GetName() string {
	return "featureFlags"
}

func (f features) GetTag() string {
	return "ff"
}

type segments struct{}

func (s segments) GetName() string {
	return "segments"
}

func (s segments) GetTag() string {
	return "seg"
}

type datatests struct{}

func (d datatests) GetName() string {
	return "datatests"
}

func (d datatests) GetTag() string {
	return "test"
}

var (
	Features  interfaces.Category = features{}
	Segments  interfaces.Category = segments{}
	Datatests interfaces.Category = datatests{}
	AllCats                       = []interfaces.Category{Features, Segments, Datatests}
)
