# Go server Side SDK

## Introduction

This is the Go Server-Side SDK for the 100% open-source feature flags management
platform [FeatBit](https://github.com/featbit/featbit). It is intended for use in a multi-user Go
server applications.

This SDK has two main purposes:

- Store the available feature flags and evaluate the feature flag variation for a given user
- Send feature flag usage and custom events for the insights and A/B/n testing.

## Data synchonization

We use websocket to make the local data synchronized with the FeatBit server, and then store them in memory by
default. Whenever there is any change to a feature flag or its related data, this change will be pushed to the SDK and
the average synchronization time is less than 100 ms. Be aware the websocket connection may be interrupted due to
internet outage, but it will be resumed automatically once the problem is gone.

If you want to use your own data source, see [Offline](#offline).

## Get Started

Go Server Side SDK is based on go 1.13, so you need to install go 1.13 or above.

### Installation

```
go get github.com/featbit/featbit-go-sdk
```

### Basic usage
```go

envSecret := "<replace-with-your-env-secret>"
streamingUrl := "ws://localhost:5100"
eventUrl := "http://localhost:5100"

client, err := featbit.NewFBClient(envSecret, streamingUrl, eventUrl)

// should close the client when you don't need it anymore
defer func() {
    if client != nil {
        _ = client.Close()
    }
}()

if err == nil && client.IsInitialized() {
    user, _ := interfaces
	    .NewUserBuilder("<replace-with-your-user-key>")
	    .UserName("<replace-with-your-user-name>")
	    .Build()
    flagValue, Detail, _ := client.Variation("<replace-with-your-flag-key>", user, false)
    fmt.Printf("flag %s, returns %s for user %s \n", ed.KeyName, ed.Variation, user.GetKey())
    fmt.Printf("Reason Description: %s \n", ed.Reason)
}

```
Note that the _**envSecret**_, _**streamUrl**_ and _**eventUrl**_ are required to initialize the SDK.

### Examples

- [Go Demo](https://github.com/featbit/featbit-samples/blob/main/samples/dino-game/demo-golang/go_demo.go)

### FBClient

Applications SHOULD instantiate a single instance for the lifetime of the application. In the case where an application
needs to evaluate feature flags from different environments, you may create multiple clients, but they should still be
retained for the lifetime of the application rather than created per request or per thread.

### Bootstrapping

The bootstrapping is in fact the call of constructor of `featbit.FBClient`, in which the SDK will be initialized, using
streaming from your feature management platform.

The constructor will return when it successfully connects, or when the timeout set
by `featbit.FBConfig.StartWait`
(default: 15 seconds) expires, whichever comes first. If it has not succeeded in connecting when the timeout elapses,
you will receive the client in an uninitialized state where feature flags will return default values; it will still
continue trying to connect in the background unless there has been an `net.DNSError` or you close the
client. You can detect whether initialization has succeeded by calling `featbit.FBClient.IsInitialized()`.

```go
config := featbit.FBConfig{StartWait: 10 * time.Second}
client, err := featbit.MakeCustomFBClient(envSecret, streamingUrl, eventUrl, config)
// DO not forget to close the client when you don't need it anymore
if err ==nil && !client.IsInitialized() {
// do whatever is appropriate if initialization has timed out
}

```

If you prefer to have the constructor return immediately, and then wait for initialization to finish at some other
point, you can use `featbit.FBClient.GetDataUpdateStatusProvider()`, which provides an asynchronous way, as follows:

```go
config := featbit.FBConfig{StartWait: 0}
client, err := featbit.MakeCustomFBClient(envSecret, streamingUrl, eventUrl, config)
// DO not forget to close the client when you don't need it anymore
// later...
if err !=nil {
    return
}
ok := client.GetDataSourceStatusProvider().WaitForOKState(10 * time.Second)
if !ok {
// do whatever is appropriate if initialization has timed out
}

```

### FBClient, FBConfig and Components

In the most case, you don't need to care about `featbit.FBConfig` and the internal components, just initialize SDK like:

```go

client, err := featbit.NewFBClient(envSecret, streamingUrl, eventUrl)

```

`envSecret` _**sdkKey(envSecret)**_ is id of your project in FeatBit feature flag center

`streamingURL`: URL of your feature management platform to synchronize feature flags, user segments, etc.

`eventURL`: URL of your feature management platform to send analytics events

`StartWait`: how long the constructor will block awaiting a successful data sync. Setting this to a zero or negative
duration will not block and cause the constructor to return immediately.

`Offline`: Set whether SDK is offline. when set to true no connection to your feature management platform anymore

`featbit.FBConfig` provides advanced configuration options for setting the SDK component, or you want to customize the behavior
of build-in components.

`NetworkFactory`: sets the SDK networking configuration, _**DO NOT**_ change it unless you should set some advanced configuration such as
HTTP Proxy, TLS etc.

`factories.NetworkBuilder` is the default `NetworkFactory`

```go

factory := factories.NewNetworkBuilder()
factory.ProxyUrl("http://username:password@146.137.9.45:65233")

config := featbit.DefaultFBConfig
config.NetworkFactory = factory
client, err := featbit.MakeCustomFBClient(envSecret, streamingUrl, eventUrl, *config)
// or
config := featbit.FBConfig{NetworkFactory: factory}
client, err := featbit.MakeCustomFBClient(envSecret, streamingUrl, eventUrl, config)

```

`DataStorageFactory` sets the implementation of `interfaces.DataStorage` to be used for holding feature flags and
related data received from feature flag center SDK sets the implementation of the data storage, in using `factories.InMemoryStorageBuilder` by default
to instantiate a memory data storage. Developers can customize the data storage to persist received data in redis,
mongodb, etc.

`DataSynchronizerFactory` SDK sets the implementation of the `interfaces.DataSynchronizer` that receives feature flag data
from feature flag center, in using `factories.StreamingBuilder` by default
If Developers would like to know what the implementation is, they can read the GoDoc and source code.

`InsightProcessorFactory` SDK which sets the implementation of `interfaces.InsightProcessor` to be used for processing analytics events.
using a factory object. The default is `factories.InsightProcessorBuilder`.
If Developers would like to know what the implementation is, they can read the GoDoc and source code.

It's not recommended to change the default factories in the `featbit.FBConfig`

### Offline

In the offline mode, SDK DOES not exchange any data with your feature management platform

To open the offline mode:

```go
config := featbit.DefaultFBConfig
featbit.Offline = true
featbit.StartWait = 1 * time.Millisecond
client, err := featbit.MakeCustomFBClient(envSecret, streamingUrl, eventUrl, *config)
// or
config := FBConfig{Offline: true, StartWait: 1 * time.Millisecond}
client, err := featbit.MakeCustomFBClient(envSecret, streamingUrl, eventUrl, config)

```

When you put the SDK in offline mode, no insight message is sent to the server and all feature flag evaluations return
fallback values because there are no feature flags or segments available. If you want to use your own data source,
SDK allows users to populate feature flags and segments data from a JSON string.

Here is an example: [fbclient_test_data.json](fixtures/fbclient_test_data.json).

The format of the data in flags and segments is defined by FeatBit and is subject to change. Rather than trying to
construct these objects yourself, it's simpler to request existing flags directly from the FeatBit server in JSON format
and use this output as the starting point for your file. Here's how:

```shell
# replace http://localhost:5100 with your evaluation server url
curl -H "Authorization: <your-env-secret>" http://localhost:5100/api/public/sdk/server/latest-all > featbit-bootstrap.json
```

Then you can use this file to initialize the SDK in offline mode:
    
```go
// first load data from file and then 
ok, _ := client.InitializeFromExternalJson(string(jsonBytes))
```

### FBUser

`FBUser`: A collection of attributes that can affect flag evaluation, usually corresponding to a user of your
application.
This object contains built-in properties(`key`, `userName`). The `key` and `userName` are required.
The `key` must uniquely identify each user; this could be a username or email address for authenticated users, or an ID
for anonymous users.
The `userName` is used to search your user quickly.
You may also define custom properties with arbitrary names and values.

```go
// FBUser creation
user, err := NewUserBuilder("key").UserName("name").Custom("property", "value").Build()
```

### Evaluation

SDK calculates the value of a feature flag for a given user, and returns a flag value and `interfaces.EvalDetail` that describes the way
that the value was determined.

SDK will initialize all the related data(feature flags, segments etc.) in the bootstrapping and receive the data updates
in real time, as mentioned in [Bootstrapping](#bootstrapping).

After initialization, the SDK has all the feature flags in the memory and all evaluation is done _**locally and
synchronously**_, the average evaluation time is < _**10**_ ms.

If evaluation called before Go SDK client initialized, or you set the wrong flag key or user for the evaluation, SDK will return
the default value you set.

SDK supports String, Boolean, and Number and Json as the return type of flag values:

- Variation(for string)
- BoolVariation
- IntVariation
- DoubleVariation
- JsonVariation

```go
// be sure that SDK is initialized before evaluation
// DO not forget to close client when you are done with it
if client.isInitialized() {
    // Flag value
    // returns a string variation
    variation, detail, _ := client.Variation("flag key", user, "Not Found")
}
```

`featbit.FBClient.AllLatestFlagsVariations(user)` returns all variations for a given user. You can retrieve the flag value or details
for a specific flag key:

- GetStringVariation
- GetBoolVariation
- GetIntVariation
- GetDoubleVariation
- GetJsonVariation

```go
// be sure that SDK is initialized before evaluation
// DO not forget to close client when you are done with it
if client.isInitialized() {
    // get all variations for a given user in your project 
    allState, _ := client.AllLatestFlagsVariations(user)
    variation, detail, _ := allState.GetStringVariation("flag key", "Not Found")
}
```

### Experiments (A/B/n Testing)

We support automatic experiments for pageviews and clicks, you just need to set your experiment on FeatBit platform,
then you should be able to see the result in near real time after the experiment is started.

In case you need more control over the experiment data sent to our server, we offer a method to send custom event.

```go
// for the percentage experiment
client.TrackPercentageMetric(user, eventName)
// for the numeric experiment
client.TrackNumericMetric(user, eventName, numericValue)
```

Make sure `featbit.FBClient.TrackPercentageMetric()` or `featbit.FBClient.TrackNumericMetric()`  is called after the related feature flag is called,
otherwise the custom event may not be included into the experiment result.


## Getting support

- If you have a specific question about using this sdk, we encourage you
  to [ask it in our slack](https://join.slack.com/t/featbit/shared_invite/zt-1ew5e2vbb-x6Apan1xZOaYMnFzqZkGNQ).
- If you encounter a bug or would like to request a
  feature, [submit an issue](https://github.com/featbit/dotnet-server-sdk/issues/new).

## See Also
- [Connect To Go Sdk](https://docs.featbit.co/docs/getting-started/4.-connect-an-sdk/server-side-sdks/go-sdk)
