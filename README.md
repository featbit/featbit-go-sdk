# Go server Side SDK

## Introduction

This is the Go Server Side SDK for the feature management platform FeatBit. It is intended for use in a multi-user Go
server applications.

This SDK has two main purposes:

- Store the available feature flags and evaluate the feature flag variation for a given user
- Send feature flag usage and custom events for the insights and A/B/n testing.

## Data synchonization

We use websocket to make the local data synchronized with the server, and then store them in memory by default.Whenever
there is any change to a feature flag or its related data, this change will be pushed to the SDK, the average
synchronization time is less than **100** ms. Be aware the websocket connection may be interrupted due to internet
outage, but it will be resumed automatically once the problem is gone.

## Offline mode support

In the offline mode, SDK DOES not exchange any data with your feature management platform

In the following situation, the SDK would work when there is no internet connection: it has been initialized in
using `featbit.FBClient.InitializeFromExternalJson()`

To open the offline mode:

```go
config := featbit.DefaultFBConfig
featbit.Offline = true

// or

config := featbit.FBConfig{Offline: true}

```

## Evaluation of a feature flag

SDK will initialize all the related data(feature flags, segments etc.) in the bootstrapping and receive the data updates
in real time, as mentioned in the above.

After initialization, the SDK has all the feature flags in the memory and all evaluation is done locally and
synchronously, the average evaluation time is < **10** ms.

## Installation

```
go get github.com/featbit/featbit-go-sdk
```

## SDK

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

client, _ := featbit.NewFBClient(envSecret, streamingUrl, eventUrl)
if !client.IsInitialized() {
// do whatever is appropriate if initialization has timed out
}

```

If you prefer to have the constructor return immediately, and then wait for initialization to finish at some other
point, you can use `featbit.FBClient.GetDataUpdateStatusProvider()`, which provides an asynchronous way, as follows:

```go
config := featbit.FBConfig{StartWait: 0}
client, _ := featbit.MakeCustomFBClient(envSecret, streamingUrl, eventUrl, config)
// later...
ok := client.GetDataSourceStatusProvider().WaitForOKState(10 * time.Second)
if !ok {
// do whatever is appropriate if initialization has timed out
}

```

Note that the _**sdkKey(envSecret)**_ is mandatory.

### FBClient, FBConfig and Components

In the most case, you don't need to care about `featbit.FBConfig` and the internal components, just initialize SDK like:

```go

client, _ := featbit.NewFBClient(envSecret, streamingUrl, eventUrl)

```

`envSecret` _**sdkKey(envSecret)**_ is id of your project in FeatBit feature flag center

`streamingURL`: URL of your feature management platform to synchronise feature flags, user segments, etc.

`eventURL`: URL of your feature management platform to send analytics events

If you would like to run in the offline mode or change the timeout:

```go
config := featbit.DefaultFBConfig
featbit.Offline = true
featbit.StartWait = 0

// or
config := featbit.FBConfig{StartWait: 0, Offline: true}

client, _ := featbit.MakeCustomFBClient(envSecret, streamingUrl, eventUrl, config)

```

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
// or
config := featbit.FBConfig{NetworkFactory: factory}

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

### Evaluation

SDK calculates the value of a feature flag for a given user, and returns a flag value/an object that describes the way
that the value was determined.

`FBUser`: A collection of attributes that can affect flag evaluation, usually corresponding to a user of your
application.
This object contains built-in properties(`key`, `userName`). The `key` and `userName` are required.
The `key` must uniquely identify each user; this could be a username or email address for authenticated users, or a ID
for anonymous users.
The `userName` is used to search your user quickly.
You may also define custom properties with arbitrary names and values.

```go
client, _ := featbit.NewFBClient(envSecret, streamingUrl, eventUrl)

// FBUser creation
user, _ := NewUserBuilder("key").UserName("name").Custom("property", "value").Build()

// be sure that SDK is initialized
// this is not required
if(client.isInitialized()){
// Flag value
// returns a string variation
variation, detail, _ := client.Variation("flag key", user, "Not Found");

// get all variations for a given user in your project
AllFlagStates states = client.AllLatestFlagsVariations(user);
variation, detail, _  = states.GetStringVariation("flag key", user, "Not Found");
}
```

If evaluation called before Go SDK client initialized, or you set the wrong flag key or user for the evaluation, SDK
will return the default value you set.

SDK supports String, Boolean, and Number and Json as the return type of flag values, see GoDocs for more details.

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


