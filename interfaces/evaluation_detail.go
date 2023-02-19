package interfaces

// EvalDetail is an interface combining the result of a flag evaluation with an explanation of how it was calculated.
type EvalDetail struct {

	// Variation is the result of the flag evaluation. This will be either one of the flag's variations or
	// the default value that was passed to the Variation.
	//
	// The type of the result should be string, int, float64, bool or json.
	Variation interface{} `json:"variation"`

	// Reason describes the main factor that influenced the flag
	Reason string `json:"reason"`
	// GetKeyName returns the feature key of the latest evaluated feature flag
	KeyName string `json:"keyName"`
	// GetName returns the name of the latest evaluated feature flag
	Name string `json:"name"`
}

type AllFlagState interface {
	IsSuccess() bool
	Reason() string
	GetStringVariation(featureFlagKey string, defaultValue string) (string, EvalDetail, error)
	GetBoolVariation(featureFlagKey string, defaultValue bool) (bool, EvalDetail, error)
	GetIntVariation(featureFlagKey string, defaultValue int) (int, EvalDetail, error)
	GetDoubleVariation(featureFlagKey string, defaultValue float64) (float64, EvalDetail, error)
	GetJsonVariation(featureFlagKey string, defaultValue interface{}) (interface{}, EvalDetail, error)
}
