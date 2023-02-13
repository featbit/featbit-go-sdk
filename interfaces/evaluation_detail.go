package interfaces

// EvalDetail is an interface combining the result of a flag evaluation with an explanation of how it was calculated.
type EvalDetail interface {
	// GetVariation is the result of the flag evaluation. This will be either one of the flag's variations or
	// the default value that was passed to the Variation method.
	//
	// The type of the result should be string, int, float64, bool or json.
	GetVariation() interface{}

	// GetReason describes the main factor that influenced the flag
	GetReason() string
	// GetKeyName returns the feature key of the latest evaluated feature flag
	GetKeyName() string
	// GetName returns the name of the latest evaluated feature flag
	GetName() string
}
