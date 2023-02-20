package interfaces

import "io"

// Sender interface for the http connection to help FeatBit API send or receive the details of feature flags, user segments, events etc.
type Sender interface {
	io.Closer
	// PostJson sends the json objects to feature flag center in the post method
	PostJson(uri string, bytes []byte) ([]byte, error)
}
