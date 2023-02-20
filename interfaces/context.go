package interfaces

// BasicConfig specifies the basic configurations of SDK that will be used for all components
type BasicConfig interface {
	// GetEnvSecret returns the env secret
	GetEnvSecret() string
	// GetStreamingUri returns the streaming url
	GetStreamingUri() string
	// GetEventUri return the event url
	GetEventUri() string
}

// Context is used to create components, context information provided by the FeatBit GO SDK
// This is passed as parameter to component factories. Component factories do not receive the entire FeatBit config
// because it contains only factory implementations.
//
// Note that the actual implementation class may contain other properties that are only relevant to the built-in
// SDK components and are therefore not part of the public interface; this allows the SDK to add its own
// context information as needed without disturbing the public API.
type Context interface {
	BasicConfig
	// GetNetwork returns the FeatBit network, such as HTTP client, Websocket client
	GetNetwork() Network
}
