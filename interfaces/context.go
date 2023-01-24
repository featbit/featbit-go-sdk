package interfaces

type BasicConfig interface {
	GetEnvSecret() string
	GetStreamingUri() string
	GetEventUri() string
}

type Context interface {
	BasicConfig
	GetNetwork() Network
}
