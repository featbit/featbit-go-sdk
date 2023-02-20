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

// AllFlagState provides a standard return responding the request of getting all flag values from SDK
type AllFlagState interface {
	// IsSuccess returns true if the last evaluation is successful
	IsSuccess() bool
	// Reason return `OK` if the last evaluation is successful, otherwise return the reason
	Reason() string
	// GetStringVariation return the string value of a given flag key name or default value if flag not existed;
	// details that that explains how the flag value is explained and the error if any.
	//
	// The method sends insight events back to feature flag center
	GetStringVariation(featureFlagKey string, defaultValue string) (string, EvalDetail, error)
	// GetBoolVariation return the bool value of a given flag key name or default value if flag not existed;
	// details that that explains how the flag value is explained and the error if any.
	//
	// The method sends insight events back to feature flag center
	GetBoolVariation(featureFlagKey string, defaultValue bool) (bool, EvalDetail, error)
	// GetIntVariation return the int value of a given flag key name or default value if flag not existed;
	// details that that explains how the flag value is explained and the error if any.
	//
	// The method sends insight events back to feature flag center
	GetIntVariation(featureFlagKey string, defaultValue int) (int, EvalDetail, error)
	// GetDoubleVariation return the float value of a given flag key name or default value if flag not existed;
	// details that that explains how the flag value is explained and the error if any.
	//
	// The method sends insight events back to feature flag center
	GetDoubleVariation(featureFlagKey string, defaultValue float64) (float64, EvalDetail, error)
	// GetJsonVariation return the json object of a given flag key name or default value if flag not existed;
	// details that that explains how the flag value is explained and the error if any.
	//
	// The method sends insight events back to feature flag center
	GetJsonVariation(featureFlagKey string, defaultValue interface{}) (interface{}, EvalDetail, error)
}
