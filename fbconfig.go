package featbit

import (
	"github.com/featbit/featbit-go-sdk/v2/factories"
	. "github.com/featbit/featbit-go-sdk/v2/interfaces"
	"log/slog"
	"time"
)

// LevelTrace is the level used for verbose SDK diagnostics. A custom handler
// must enable this level to receive trace records.
const LevelTrace = slog.Level(-8)

const (
	INFO = iota
	WARN
	ERROR

	TRACE = -2
	DEBUG = -1
)

// FBConfig exposes advanced configuration options for the FBClient
//		config = FBConfig{Offline: true}
type FBConfig struct {
	// Offline whether SDK is offline
	Offline bool
	// StartWait how long the constructor will block awaiting a successful data sync
	//
	// Setting this to a zero or negative duration will not block and cause the constructor to return immediately.
	StartWait time.Duration
	// NetworkFactory a factory object which sets the SDK networking configuration Depending on the implementation,
	// the factory may be a builder that allows you to set other configuration options as well.
	NetworkFactory NetworkFactory
	// DataStorageFactory a factory object which sets the implementation of interfaces.DataStorage to be used for holding feature flags and
	// related data received from feature flag center. Depending on the implementation, the factory may be a builder that
	// allows you to set other configuration options as well.
	DataStorageFactory DataStorageFactory
	// DataSynchronizerFactory a factory object which sets the implementation of the interfaces.DataSynchronizer that receives feature flag data
	// from feature flag center.
	//
	// Depending on the implementation, the factory may be a builder that allows you to set other configuration options as well.
	DataSynchronizerFactory DataSynchronizerFactory
	// InsightProcessorFactory a factory object which sets the implementation of interfaces.InsightProcessor to be used for processing analytics events.
	//
	// Depending on the implementation, the factory may be a builder that allows you to set other configuration options as well.
	InsightProcessorFactory InsightProcessorFactory
	// Logger receives this client's logs, including logs from background components.
	// Its handler controls filtering, formatting and output and must support concurrent calls.
	// The SDK does not change the global slog logger or close the supplied logger.
	// If nil, the SDK creates a text logger writing to stdout using LogLevel.
	Logger *slog.Logger
	// LogLevel controls the default logger using TRACE, DEBUG, INFO, WARN or ERROR.
	// It is ignored when Logger is provided; these integers are not slog.Level values.
	LogLevel int
}

// DefaultFBConfig FeatBit default configuration
var DefaultFBConfig *FBConfig = &FBConfig{
	Offline:                 false,
	StartWait:               15 * time.Second,
	NetworkFactory:          factories.NewNetworkBuilder(),
	DataStorageFactory:      factories.NewInMemoryStorageBuilder(),
	DataSynchronizerFactory: factories.NewStreamingBuilder(),
	InsightProcessorFactory: factories.NewInsightProcessorBuilder(),
	LogLevel:                INFO,
}
