package featbit

import (
	"encoding/json"
	"fmt"
	"github.com/featbit/featbit-go-sdk/factories"
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal"
	"github.com/featbit/featbit-go-sdk/internal/datasynchronization"
	"github.com/featbit/featbit-go-sdk/internal/dataupdating"
	"github.com/featbit/featbit-go-sdk/internal/types/data"
	"github.com/featbit/featbit-go-sdk/internal/types/insight"
	"github.com/featbit/featbit-go-sdk/internal/util"
	"github.com/featbit/featbit-go-sdk/internal/util/log"
	"sync"
	"time"
)

type FBClient struct {
	offline                  bool
	dataStorage              DataStorage
	dataSynchronizer         DataSynchronizer
	dataUpdater              DataUpdater
	dataUpdateStatusProvider DataUpdateStatusProvider
	insightProcessor         InsightProcessor
	evaluator                *evaluator
	getFlag                  func(key string) *data.FeatureFlag
	sendEvent                func(Event)
}

var (
	envSecretInvalid      = fmt.Errorf("invalid env secret")
	hostInvalid           = fmt.Errorf("invalid streaming url or event url")
	initializationTimeout = fmt.Errorf("timeout encountered waiting for client initialization")
	initializationFailed  = fmt.Errorf("client initialization failed")
	clientNotInitialized  = fmt.Errorf("evaluation is called before client is initialized")
	emptyClient           = fmt.Errorf("empty client, please call constructor")
	flagNotFound          = fmt.Errorf("feature flag not found")
	userInvalid           = fmt.Errorf("invalid user")
	evalFailed            = fmt.Errorf("evaluation failed")
	evalWrongType         = fmt.Errorf("flag type doesn't match the request")
)

// NewFBClient creates a new client instance that connects to your feature flag center with the default configuration.
// For advanced configuration options, use MakeCustomFBClient. Calling NewFBClient is exactly equivalent to
// calling MakeCustomClient with the config parameter set to a default value.
//
// Unless it is configured to be offline with FBConfig.Offline, the client will begin attempting to connect to feature flag center as soon as you call this constructor.
//The constructor will return when it successfully connects, or when the timeout set by the FBConfig.StartWait parameter expires, whichever comes first.
//
// If the timeout(15s) elapsed without a successful connection, it still returns a client instance-- in an initializing state,
// where feature flags will return default values-- and the error value is initializationTimeout. In this case, it will still continue trying to connect in the background.
//
// If there was an unexpected error such that it cannot succeed by retrying-- for instance, the envSecret key is
// invalid or an DNS error-- it will return a client instance in an uninitialized state, and the error value is initializationFailed.
//
// The way to monitor the client's status, use FBClient.IsInitialized or FBClient.GetDataUpdateStatusProvider.
//
//     client, _ := featbit.NewFBClient(envSecret, streamingUrl, eventUrl)
//
//     if !client.IsInitialized() {
//         // do whatever is appropriate if initialization has timed out
//     }
//
// If you set FBConfig.StartWait to zero, the function will return immediately after creating the client instance, and do any further initialization in the background.
//
//     client, _ := featbit.MakeCustomFBClient(envSecret, streamingUrl, eventUrl, config)
//
//     // later...
//     ok := client.GetDataSourceStatusProvider().WaitForOKState(10 * time.Second)
//     if !ok {
//         // do whatever is appropriate if initialization has timed out
//     }
//
// The only time it returns nil instead of a client instance is if the client cannot be created at all due to
// an invalid configuration. This is rare, but could happen if for example you specified a custom TLS
// certificate file that did not load a valid certificate, you inputted an invalid env secret key, etc...
func NewFBClient(envSecret string, streamingUrl string, eventUrl string) (*FBClient, error) {
	config := DefaultFBConfig
	return MakeCustomFBClient(envSecret, streamingUrl, eventUrl, *config)
}

// MakeCustomFBClient creates a new client instance that connects to your feature flag center with the custom configuration.
//
// The FBConfig allows customization of all SDK properties; some of these are represented directly as
// fields in FBConfig, while others are set by builder methods on a more specific configuration object. See FBConfig for details.
//
// Unless it is configured to be offline with FBConfig.Offline, the client will begin attempting to connect to feature flag center as soon as you call this constructor.
//The constructor will return when it successfully connects, or when the timeout set by the FBConfig.StartWait parameter expires, whichever comes first.
//
// If the timeout(15s) elapsed without a successful connection, it still returns a client instance-- in an initializing state,
// where feature flags will return default values-- and the error value is initializationTimeout. In this case, it will still continue trying to connect in the background.
//
// If there was an unexpected error such that it cannot succeed by retrying-- for instance, the envSecret key is
// invalid or an DNS error-- it will return a client instance in an uninitialized state, and the error value is initializationFailed.
//
// The way to monitor the client's status, use FBClient.IsInitialized or FBClient.GetDataUpdateStatusProvider.
//
//     client, _ := featbit.MakeCustomFBClient(envSecret, streamingUrl, eventUrl, config)
//
//     if !client.IsInitialized() {
//         // do whatever is appropriate if initialization has timed out
//     }
//
// If you set FBConfig.StartWait to zero, the function will return immediately after creating the client instance, and do any further initialization in the background.
//
//     client, _ := featbit.MakeCustomFBClient(envSecret, streamingUrl, eventUrl, config)
//
//     // later...
//     ok := client.GetDataSourceStatusProvider().WaitForOKState(10 * time.Second)
//     if !ok {
//         // do whatever is appropriate if initialization has timed out
//     }
//
// The only time it returns nil instead of a client instance is if the client cannot be created at all due to
// an invalid configuration. This is rare, but could happen if for example you specified a custom TLS
// certificate file that did not load a valid certificate, you inputted an invalid env secret key, etc...
func MakeCustomFBClient(envSecret string, streamingUrl string, eventUrl string, config FBConfig) (*FBClient, error) {
	logger := &log.SimpleLogger{Level: config.LogLevel}
	log.SetLogger(logger)
	if !config.Offline {
		if !util.IsEnvSecretValid(envSecret) {
			return nil, envSecretInvalid
		} else if !util.IsUrl(streamingUrl) || !util.IsUrl(eventUrl) {
			return nil, hostInvalid
		}
	} else {
		log.LogInfo("FB GO SDK: SDK is in offline mode")
	}
	networkFactory := config.NetworkFactory
	if networkFactory == nil {
		networkFactory = factories.NewNetworkBuilder()
	}
	ctx, err := internal.FromConfig(envSecret, streamingUrl, eventUrl, networkFactory)
	if err != nil {
		return nil, err
	}
	client := &FBClient{offline: config.Offline}
	// init components
	// data storage
	dataStorageFactory := config.DataStorageFactory
	if dataStorageFactory == nil {
		dataStorageFactory = factories.NewInMemoryStorageBuilder()
	}
	client.dataStorage, err = dataStorageFactory.CreateDataStorage(ctx)
	if err != nil {
		return nil, err
	}
	//evaluator
	client.getFlag = func(key string) *data.FeatureFlag {
		if item, e := client.dataStorage.Get(data.Features, key); e == nil {
			if flag, ok := item.(*data.FeatureFlag); ok {
				return flag
			}
		}
		return nil
	}

	getSegment := func(key string) *data.Segment {
		if item, e := client.dataStorage.Get(data.Segments, key); e == nil {
			if segment, ok := item.(*data.Segment); ok {
				return segment
			}
		}
		return nil
	}
	client.evaluator = newEvaluator(client.getFlag, getSegment)

	// data updater
	dataUpdater := dataupdating.NewDataUpdaterImpl(client.dataStorage)
	client.dataUpdater = dataUpdater
	// data update status provider
	client.dataUpdateStatusProvider = dataupdating.NewDataUpdateStatusProviderImpl(dataUpdater)

	// run insight processor
	insightProcessorFactory := config.InsightProcessorFactory
	if client.offline {
		insightProcessorFactory = factories.ExternalEventTrack()
	} else if insightProcessorFactory == nil {
		insightProcessorFactory = factories.NewInsightProcessorBuilder()
	}
	client.insightProcessor, err = insightProcessorFactory.CreateInsightProcessor(ctx)
	if err != nil {
		return nil, err
	}

	client.sendEvent = func(event Event) {
		client.insightProcessor.Send(event)
	}

	// run data synchronizer
	dataSynchronizerFactory := config.DataSynchronizerFactory
	if client.offline {
		dataSynchronizerFactory = factories.ExternalDataSynchronization()
	} else if dataSynchronizerFactory == nil {
		dataSynchronizerFactory = factories.NewStreamingBuilder()
	}
	client.dataSynchronizer, err = dataSynchronizerFactory.CreateDataSynchronizer(ctx, dataUpdater)
	if err != nil {
		return nil, err
	}
	ready := client.dataSynchronizer.Start()
	if config.StartWait > 0 {
		if _, ok := client.dataSynchronizer.(*datasynchronization.NullDataSynchronizer); !ok {
			log.LogInfo("FB GO SDK: waiting for Client initialization in %d milliseconds", config.StartWait/time.Millisecond)
		}
		select {
		case <-ready:
			if !client.dataUpdater.StorageInitialized() && !config.Offline {
				log.LogWarn("FB GO SDK: SDK just returns default variation because of no data found in the given environment")
			}
			if !client.dataSynchronizer.IsInitialized() {
				log.LogWarn("FB GO SDK: SDK was not successfully initialized")
				return client, initializationFailed
			}
			return client, nil
		case <-time.After(config.StartWait):
			log.LogWarn("FB GO SDK: timeout encountered when waiting for data update")
			// it's rare, but prevent to block data synchronizer without waiting for termination of initialization
			go func() { <-ready }()
			return client, initializationTimeout
		}

	}
	log.LogInfo("FB GO SDK: SDK starts in asynchronous mode")
	go func() { <-ready }()
	return client, nil
}

// IsInitialized tests whether the client is ready to be used.
// return true if the client is ready, or false if it is still initializing.
//
// If this value is true, it means the FBClient has succeeded at some point in connecting to feature flag center and
// has received feature flag data. It could still have encountered a connection problem after that point, so
// this does not guarantee that the flags are up-to-date; if you need to know its status in more Detail, use FBClient.GetDataUpdateStatusProvider.
//
// If this value is false, it means the client has not yet connected to feature flag center, or has permanently
// failed. In this state, feature flag evaluations will always return default values. You can use FBClient.GetDataUpdateStatusProvider
// to get current status of the client.

func (client *FBClient) IsInitialized() bool {
	if client.dataSynchronizer == nil {
		return false
	}
	return client.dataSynchronizer.IsInitialized()
}

// Close shuts down the FBClient. After calling this, the FBClient should no longer be used.
// The method will block until all pending events (if any) been sent.
func (client *FBClient) Close() error {
	log.LogInfo("FB GO SDK: Java SDK client is closing")
	if client.dataStorage != nil {
		_ = client.dataStorage.Close()
	}
	if client.dataUpdateStatusProvider != nil {
		_ = client.dataUpdateStatusProvider.Close()
	}
	if client.dataSynchronizer != nil {
		_ = client.dataSynchronizer.Close()
	}
	if client.insightProcessor != nil {
		_ = client.insightProcessor.Close()
	}
	return nil
}

// GetDataUpdateStatusProvider returns an interface for tracking the status of the interfaces.DataSynchronizer.
//
// The data synchronizer is the component that the SDK uses to get feature flags, segments such as a
// streaming connection. The interfaces.DataUpdateStatusProvider has methods
// for checking whether the interfaces.DataSynchronizer is currently operational and tracking changes in this status.
//
// The interfaces.DataUpdateStatusProvider is recommended to use when SDK starts in asynchronous mode
func (client *FBClient) GetDataUpdateStatusProvider() DataUpdateStatusProvider {
	return client.dataUpdateStatusProvider
}

// IsFlagKnown returns true if feature flag is registered in the feature flag center,
// false if any error or flag is not existed
func (client *FBClient) IsFlagKnown(featureFlagKey string) bool {
	if client.getFlag != nil {
		return client.getFlag(featureFlagKey) != nil
	}
	return false
}

// Identify register a FBUser
func (client *FBClient) Identify(user FBUser) error {
	if client.insightProcessor == nil {
		return emptyClient
	}
	eventUser := insight.ConvertFBUserToEventUser(&user)
	event := insight.NewUserEvent(eventUser)
	client.sendEvent(event)
	return nil
}

// TrackPercentageMetric reports that a user has performed an event, and associates it with a default value.
// This value is used by the experimentation feature in percentage custom metrics.
//
// The eventName normally corresponds to the event Name of a metric that you have created through the
// experiment dashboard in the feature flag center
func (client *FBClient) TrackPercentageMetric(user FBUser, eventName string) error {
	return client.TrackNumericMetric(user, eventName, 1)
}

// TrackNumericMetric reports that a user has performed an event, and associates it with a metric value.
// This value is used by the experimentation feature in numeric custom metrics.
//
// The eventName normally corresponds to the event Name of a metric that you have created through the
// experiment dashboard in the feature flag center
func (client *FBClient) TrackNumericMetric(user FBUser, eventName string, metricValue float64) error {
	if client.insightProcessor == nil {
		return emptyClient
	}
	eventUser := insight.ConvertFBUserToEventUser(&user)
	metric := insight.NewMetric(eventName, metricValue)
	event := insight.NewMetricEvent(eventUser)
	event.Add(metric)
	client.sendEvent(event)
	return nil
}

// TrackPercentageMetrics reports that a user tracks that a user performed a series of events with default values.
// These values are used by the experimentation feature in percentage custom metrics.
//
// The eventName normally corresponds to the event Name of a metric that you have created through the
// experiment dashboard in the feature flag center
func (client *FBClient) TrackPercentageMetrics(user FBUser, eventNames ...string) error {
	if client.insightProcessor == nil {
		return emptyClient
	}
	if len(eventNames) > 0 {
		eventUser := insight.ConvertFBUserToEventUser(&user)
		event := insight.NewMetricEvent(eventUser)
		for _, eventName := range eventNames {
			metric := insight.NewMetric(eventName, 1)
			event.Add(metric)
		}
		client.sendEvent(event)
	}
	return nil
}

// TrackNumericMetrics reports that a user tracks that a user performed a series of events with metric values.
// These values are used by the experimentation feature in numeric custom metrics.
//
// The eventName normally corresponds to the event Name of a metric that you have created through the
// experiment dashboard in the feature flag center
func (client *FBClient) TrackNumericMetrics(user FBUser, metrics map[string]float64) error {
	if client.insightProcessor == nil {
		return emptyClient
	}
	if len(metrics) > 0 {
		eventUser := insight.ConvertFBUserToEventUser(&user)
		event := insight.NewMetricEvent(eventUser)
		for eventName, metricValue := range metrics {
			metric := insight.NewMetric(eventName, metricValue)
			event.Add(metric)
		}
		client.sendEvent(event)
	}
	return nil
}

// Flush tells the FBClient that all pending events (if any) should be delivered as soon as possible.
// Flushing is asynchronous, so this method will return before it is complete.
// However, if you call Close(), events are guaranteed to be sent before that method returns.
func (client *FBClient) Flush() error {
	if client.insightProcessor == nil {
		return emptyClient
	}
	client.insightProcessor.Flush()
	return nil
}

// evaluateInternal internal use for evaluate flag value
func (client *FBClient) evaluateInternal(featureFlagKey string, user *FBUser, requiredType string) (*evalResult, error) {
	if !client.IsInitialized() {
		log.LogWarn("FB GO SDK: evaluation is called before GO SDK client is initialized for feature flag, well using the default value")
		return errorResult(ReasonClientNotReady, featureFlagKey, FlagNameUnknown), clientNotInitialized
	}
	flag := client.getFlag(featureFlagKey)
	if flag == nil {
		log.LogWarn("FB Go SDK: unknown feature flag %v; returning default value", featureFlagKey)
		return errorResult(ReasonFlagNotFound, featureFlagKey, FlagNameUnknown), flagNotFound

	}
	if !user.IsValid() {
		log.LogWarn("FB GO SDK: invalid user for feature flag %v, returning default value", featureFlagKey)
		return errorResult(ReasonUserNotSpecified, featureFlagKey, FlagNameUnknown), userInvalid
	}
	eventUser := insight.ConvertFBUserToEventUser(user)
	event := insight.NewFlagEvent(eventUser)
	er := client.evaluator.evaluate(flag, user, event)
	if !er.checkType(requiredType) {
		return errorResult(ReasonWrongType, featureFlagKey, er.name), evalWrongType
	}
	if er.success {
		client.sendEvent(event)
		return er, nil
	}
	log.LogError("FB GO SDK: unexpected error in evaluation")
	return errorResult(ReasonError, featureFlagKey, flag.Name), evalFailed
}

func (client *FBClient) evaluateDetail(featureFlagKey string, user *FBUser, requiredType string, defaultValue interface{}) (EvalDetail, error) {
	er, err := client.evaluateInternal(featureFlagKey, user, requiredType)
	if err != nil {
		return EvalDetail{Variation: defaultValue, Reason: er.reason, KeyName: er.keyName, Name: er.name}, err
	}
	return er.castVariationByFlagType(requiredType, defaultValue)
}

// Variation calculates the value of a feature flag for a given user,
// return a string variation for the given user, or defaultValue if the flag is disabled or an error occurs;
// the details that explains how the flag value is explained and the error if any.
//
// The method sends insight events back to feature flag center
func (client *FBClient) Variation(featureFlagKey string, user FBUser, defaultValue string) (string, EvalDetail, error) {
	ed, err := client.evaluateDetail(featureFlagKey, &user, FlagStringType, defaultValue)
	if err != nil {
		return defaultValue, ed, err
	}
	ret, _ := ed.Variation.(string)
	return ret, ed, nil
}

// BoolVariation calculates the value of a feature flag for a given user,
// return a bool variation for the given user, or defaultValue if the flag is disabled or an error occurs;
// the details that explains how the flag value is explained and the error if any.
//
// The method sends insight events back to feature flag center
func (client *FBClient) BoolVariation(featureFlagKey string, user FBUser, defaultValue bool) (bool, EvalDetail, error) {
	ed, err := client.evaluateDetail(featureFlagKey, &user, FlagBoolType, defaultValue)
	if err != nil {
		return defaultValue, ed, err
	}
	ret, _ := ed.Variation.(bool)
	return ret, ed, nil
}

// IntVariation calculates the value of a feature flag for a given user,
// return an int variation for the given user, or defaultValue if the flag is disabled or an error occurs;
// the details that explains how the flag value is explained and the error if any.
//
// The method sends insight events back to feature flag center
func (client *FBClient) IntVariation(featureFlagKey string, user FBUser, defaultValue int) (int, EvalDetail, error) {
	ed, err := client.evaluateDetail(featureFlagKey, &user, FlagNumericType, defaultValue)
	if err != nil {
		return defaultValue, ed, err
	}
	ret, _ := ed.Variation.(int)
	return ret, ed, nil
}

// DoubleVariation calculates the value of a feature flag for a given user,
// return a float variation for the given user, or defaultValue if the flag is disabled or an error occurs;
// the details that explains how the flag value is explained and the error if any.
//
// The method sends insight events back to feature flag center
func (client *FBClient) DoubleVariation(featureFlagKey string, user FBUser, defaultValue float64) (float64, EvalDetail, error) {
	ed, err := client.evaluateDetail(featureFlagKey, &user, FlagNumericType, defaultValue)
	if err != nil {
		return defaultValue, ed, err
	}
	ret, _ := ed.Variation.(float64)
	return ret, ed, nil
}

// JsonVariation calculates the value of a feature flag for a given user,
// return a json object variation for the given user, or defaultValue if the flag is disabled or an error occurs;
// the details that explains how the flag value is explained and the error if any.
//
// The method sends insight events back to feature flag center
func (client *FBClient) JsonVariation(featureFlagKey string, user FBUser, defaultValue interface{}) (interface{}, EvalDetail, error) {
	ed, err := client.evaluateDetail(featureFlagKey, &user, FlagJsonType, defaultValue)
	if err != nil {
		return defaultValue, ed, err
	}
	return ed.Variation, ed, nil
}

// AllLatestFlagsVariations returns a list of all feature flags value with details for a given user, including the reason
// describes the way the value was determined.
//
// The return type AllFlagState could be used as a cache that provides the flag value to a client side sdk or a front-end app.
// See more details in AllFlagState.
//
// This method does not send insight events back to feature flag center. See interfaces.AllFlagState
func (client *FBClient) AllLatestFlagsVariations(user FBUser) (AllFlagState, error) {
	if !client.IsInitialized() {
		log.LogWarn("FB GO SDK: evaluation is called before GO SDK client is initialized for feature flag, well using the default value")
		return &allFlagStateImpl{reason: ReasonClientNotReady}, clientNotInitialized
	}
	if !user.IsValid() {
		log.LogWarn("FB GO SDK: invalid user")
		return &allFlagStateImpl{reason: ReasonUserNotSpecified}, userInvalid
	}
	items, err := client.dataStorage.GetAll(data.Features)
	if err != nil {
		return &allFlagStateImpl{reason: ReasonError}, err
	}

	if len(items) == 0 {
		return &allFlagStateImpl{reason: ReasonFlagNotFound}, flagNotFound
	}

	ret := &allFlagStateImpl{}
	var once sync.Once
	for key, item := range items {
		if flag, ok := item.(*data.FeatureFlag); ok {
			eventUser := insight.ConvertFBUserToEventUser(&user)
			event := insight.NewFlagEvent(eventUser)
			er := client.evaluator.evaluate(flag, &user, event)
			if er.success {
				once.Do(func() {
					ret.success = true
					ret.reason = "OK"
					ret.states = make(map[string]map[evalResult]*insight.FlagEvent, len(items))
					ret.sendEvent = client.sendEvent
				})
				ret.states[key] = map[evalResult]*insight.FlagEvent{*er: event}
			}
		}
	}
	if !ret.success {
		log.LogError("FB GO SDK: unexpected error in evaluation")
		ret.reason = ReasonError
		return nil, evalFailed
	}
	return ret, nil
}

// InitializeFromExternalJson initializes FeatBit client in the offline mode
//
// Return false if the json can't be parsed or client is not in the offline mode
func (client *FBClient) InitializeFromExternalJson(jsonStr string) (bool, error) {
	if client.offline && jsonStr != "" {
		var all data.All
		if err := json.Unmarshal([]byte(jsonStr), &all); err != nil {
			return false, err
		}
		if all.IsProcessData() {
			d := all.Data
			if ok := client.dataUpdater.Init(d.ToStorageType(), d.GetTimestamp()); ok {
				client.dataUpdater.UpdateStatus(OKState())
				return true, nil
			}
		}
	}
	return false, nil
}
