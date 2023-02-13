package interfaces

import (
	"io"
)

// FBEvaluation defines the basic feature flag evaluation methods implemented by FBClient.
type FBEvaluation interface {
	// Variation calculates the value of a feature flag for a given user,
	// return the variation for the given user, or defaultValue if the flag is disabled or an error occurs
	Variation(featureFlagKey string, user FBUser, defaultValue string) (string, error)

	VariationDetail(featureFlagKey string, user FBUser, defaultValue string) (string, EvalDetail, error)

	BoolVariation(featureFlagKey string, user FBUser, defaultValue bool) (bool, error)

	BoolVariationDetail(featureFlagKey string, user FBUser, defaultValue bool) (bool, EvalDetail, error)

	IntVariation(featureFlagKey string, user FBUser, defaultValue int) (int, error)

	IntVariationDetail(featureFlagKey string, user FBUser, defaultValue int) (int, EvalDetail, error)

	DoubleVariation(featureFlagKey string, user FBUser, defaultValue float64) (float64, error)

	DoubleVariationDetail(featureFlagKey string, user FBUser, defaultValue float64) (float64, EvalDetail, error)

	JsonVariation(featureFlagKey string, user FBUser, defaultValue interface{}) (interface{}, error)

	JsonVariationDetail(featureFlagKey string, user FBUser, defaultValue interface{}) (interface{}, EvalDetail, error)
}

// FBInsight defines the methods implemented by FBClient that are specifically for generating analytics events.
type FBInsight interface {
	// Identify register a user
	Identify(user FBUser) error

	// TrackPercentageMetric  tracks that a user performed an event and provides a default numeric value for custom metrics.
	// this metric is normally used in percentage experiment
	TrackPercentageMetric(user FBUser, eventName string) error

	// TrackNumericMetric tracks that a user performed a series of events with default numeric value for custom metrics.
	// the metrics are normally used in percentage experiment
	TrackNumericMetric(user FBUser, eventName string, metricValue float64) error

	// TrackPercentageMetrics tracks that a user performed a series of events with default values.
	// this metric is normally used in percentage experiment
	TrackPercentageMetrics(user FBUser, eventName ...string) error

	// TrackNumericMetrics tracks that a user performed a series of events with numeric values.
	// this series of metrics is normally used in numeric experiment
	TrackNumericMetrics(user FBUser, metrics map[string]float64) error

	// Flush flushes all pending events.
	Flush() error
}

type FBClientBehaviors interface {
	io.Closer
	FBEvaluation
	FBInsight

	// IsInitialized tests whether the client is ready to be used.
	// return true if the client is ready, or false if it is still initializing
	IsInitialized() bool

	// GetDataUpdateStatusProvider returns an interface for tracking the status of the update processor.
	// The update processor is the mechanism that the SDK uses to get feature flag, such as a streaming connection.
	// DataUpdateStatusProvider is used to check whether the update processor is currently operational.
	GetDataUpdateStatusProvider() DataUpdateStatusProvider

	// IsFlagKnown returns true if the specified feature flag currently exists
	IsFlagKnown(featureFlagKey string) bool
}