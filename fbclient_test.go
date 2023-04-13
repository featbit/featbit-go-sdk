package featbit

import (
	"encoding/base64"
	"encoding/json"
	"github.com/featbit/featbit-go-sdk/factories"
	"github.com/featbit/featbit-go-sdk/fixtures"
	"github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal/datastorage"
	"github.com/featbit/featbit-go-sdk/internal/datasynchronization"
	insight2 "github.com/featbit/featbit-go-sdk/internal/insight"
	"github.com/featbit/featbit-go-sdk/internal/types/insight"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type Dummy struct {
	Code   int    `json:"code"`
	Reason string `json:"reason"`
}

var fakeEnvSecret = base64.URLEncoding.EncodeToString([]byte("http://fake"))

var testUser1, _ = interfaces.NewUserBuilder("test-user-1").Custom("country", "us").Build()
var testUser2, _ = interfaces.NewUserBuilder("test-user-2").Custom("country", "fr").Build()
var testUser3, _ = interfaces.NewUserBuilder("test-user-3").Custom("country", "cn").Custom("major", "cs").Build()
var testUser4, _ = interfaces.NewUserBuilder("test-user-4").Custom("country", "uk").Custom("major", "physics").Build()
var testUser5, _ = interfaces.NewUserBuilder("18555358000").UserName("test-user-5").Build()
var testUser6, _ = interfaces.NewUserBuilder("0603111111").UserName("test-user-6").Build()
var testUser7, _ = interfaces.NewUserBuilder("test-user-7@featbit.com").UserName("test-user-7").Build()

func TestFBClientBootStrap(t *testing.T) {
	t.Run("empty env secret", func(t *testing.T) {
		_, err := NewFBClient("", "ws://fake-url", "http://fake-url")
		assert.Equal(t, err, envSecretInvalid)
	})
	t.Run("invalid env secret", func(t *testing.T) {
		_, err := NewFBClient(fakeEnvSecret+"©öäü£", "ws://fake-url", "http://fake-url")
		assert.Equal(t, err, envSecretInvalid)
	})
	t.Run("empty url", func(t *testing.T) {
		_, err := NewFBClient(fakeEnvSecret, "", "")
		assert.Equal(t, err, hostInvalid)
	})
	t.Run("invalid url", func(t *testing.T) {
		_, err := NewFBClient(fakeEnvSecret, "urn:isbn:0-294-56559-3", "mailto:John.Doe@example.com")
		assert.Equal(t, err, hostInvalid)
	})
	t.Run("start and wait", func(t *testing.T) {
		config := FBConfig{
			StartWait:               200 * time.Millisecond,
			DataStorageFactory:      datastorage.NewMockDataStorageBuilder(),
			DataSynchronizerFactory: datasynchronization.NewMockStreamingBuilder(true, true, 100*time.Millisecond),
			InsightProcessorFactory: factories.ExternalEventTrack(),
		}
		client, err := MakeCustomFBClient(fakeEnvSecret, "ws://fake-url", "http://fake-url", config)
		require.NoError(t, err)
		assert.True(t, client.IsInitialized())
		_ = client.Close()
	})
	t.Run("start and wait but fail in initialization", func(t *testing.T) {
		config := FBConfig{
			StartWait:               200 * time.Millisecond,
			DataStorageFactory:      datastorage.NewMockDataStorageBuilder(),
			DataSynchronizerFactory: datasynchronization.NewMockStreamingBuilder(false, true, 100*time.Millisecond),
			InsightProcessorFactory: factories.ExternalEventTrack(),
		}
		client, err := MakeCustomFBClient(fakeEnvSecret, "ws://fake-url", "http://fake-url", config)
		assert.Equal(t, err, initializationFailed)
		assert.False(t, client.IsInitialized())
		res, detail, err := client.Variation("ff-test-string", testUser1, "error")
		assert.Equal(t, err, clientNotInitialized)
		assert.Equal(t, ReasonClientNotReady, detail.Reason)
		assert.Equal(t, "error", res)
		_ = client.Close()
	})
	t.Run("start and wait but no data loaded", func(t *testing.T) {
		config := FBConfig{
			StartWait:               200 * time.Millisecond,
			DataStorageFactory:      datastorage.NewMockDataStorageBuilder(),
			DataSynchronizerFactory: datasynchronization.NewMockStreamingBuilder(true, false, 100*time.Millisecond),
			InsightProcessorFactory: factories.ExternalEventTrack(),
		}
		client, err := MakeCustomFBClient(fakeEnvSecret, "ws://fake-url", "http://fake-url", config)
		require.NoError(t, err)
		assert.False(t, client.IsInitialized())
		res, detail, err := client.Variation("ff-test-string", testUser1, "error")
		assert.Equal(t, err, clientNotInitialized)
		assert.Equal(t, ReasonClientNotReady, detail.Reason)
		assert.Equal(t, "error", res)
		_ = client.Close()
	})
	t.Run("start and no wait", func(t *testing.T) {
		config := FBConfig{
			StartWait:               0,
			DataStorageFactory:      datastorage.NewMockDataStorageBuilder(),
			DataSynchronizerFactory: datasynchronization.NewMockStreamingBuilder(true, true, 100*time.Millisecond),
			InsightProcessorFactory: factories.ExternalEventTrack(),
		}
		client, err := MakeCustomFBClient(fakeEnvSecret, "ws://fake-url", "http://fake-url", config)
		require.NoError(t, err)
		assert.False(t, client.IsInitialized())
		res, detail, err := client.Variation("ff-test-string", testUser1, "error")
		assert.Equal(t, err, clientNotInitialized)
		assert.Equal(t, ReasonClientNotReady, detail.Reason)
		assert.Equal(t, "error", res)
		allState, err1 := client.AllLatestFlagsVariations(testUser1)
		assert.Equal(t, err1, clientNotInitialized)
		assert.Equal(t, ReasonClientNotReady, allState.Reason())
		assert.False(t, allState.IsSuccess())
		if client.GetDataUpdateStatusProvider().WaitForOKState(200 * time.Millisecond) {
			assert.True(t, client.IsInitialized())
			res, _, _ = client.Variation("ff-test-string", testUser1, "error")
			assert.Equal(t, "others", res)
		}
		_ = client.Close()
	})
	t.Run("start and time out", func(t *testing.T) {
		config := FBConfig{
			StartWait:               50 * time.Millisecond,
			DataStorageFactory:      datastorage.NewMockDataStorageBuilder(),
			DataSynchronizerFactory: datasynchronization.NewMockStreamingBuilder(true, true, 100*time.Millisecond),
			InsightProcessorFactory: factories.ExternalEventTrack(),
		}
		client, err := MakeCustomFBClient(fakeEnvSecret, "ws://fake-url", "http://fake-url", config)
		assert.Equal(t, err, initializationTimeout)
		assert.False(t, client.IsInitialized())
		res, detail, err := client.Variation("ff-test-string", testUser1, "error")
		assert.Equal(t, err, clientNotInitialized)
		assert.Equal(t, ReasonClientNotReady, detail.Reason)
		assert.Equal(t, "error", res)
		if client.GetDataUpdateStatusProvider().WaitForOKState(200 * time.Millisecond) {
			assert.True(t, client.IsInitialized())
			res, _, _ = client.Variation("ff-test-string", testUser1, "error")
			assert.Equal(t, "others", res)
		}
		_ = client.Close()
	})
}

func TestFBEvaluation(t *testing.T) {
	config := FBConfig{Offline: true, StartWait: 1 * time.Millisecond}
	client, _ := MakeCustomFBClient(fakeEnvSecret, "ws://fake-url", "http://fake-url", config)
	jsonBytes, _ := fixtures.LoadFBClientTestData()
	b, e := client.InitializeFromExternalJson(string(jsonBytes))
	require.NoError(t, e)
	assert.True(t, b)
	t.Run("bool variation", func(t *testing.T) {
		res, detail, err := client.BoolVariation("ff-test-bool", testUser1, false)
		require.NoError(t, err)
		assert.True(t, res)
		assert.Equal(t, ReasonTargetMatch, detail.Reason)
		res, detail, err = client.BoolVariation("ff-test-bool", testUser2, false)
		require.NoError(t, err)
		assert.True(t, res)
		assert.Equal(t, ReasonTargetMatch, detail.Reason)
		res, detail, err = client.BoolVariation("ff-test-bool", testUser3, false)
		require.NoError(t, err)
		assert.False(t, res)
		assert.Equal(t, ReasonFallthrough, detail.Reason)
		res, detail, err = client.BoolVariation("ff-test-bool", testUser4, false)
		require.NoError(t, err)
		assert.True(t, res)
		assert.Equal(t, ReasonFallthrough, detail.Reason)
	})
	t.Run("numeric variation", func(t *testing.T) {
		res, detail, err := client.IntVariation("ff-test-number", testUser1, -1)
		require.NoError(t, err)
		assert.Equal(t, 1, res)
		assert.Equal(t, ReasonRuleMatch, detail.Reason)
		res, detail, err = client.IntVariation("ff-test-number", testUser2, -1)
		require.NoError(t, err)
		assert.Equal(t, 33, res)
		assert.Equal(t, ReasonRuleMatch, detail.Reason)
		res1, detail1, err1 := client.DoubleVariation("ff-test-number", testUser3, float64(-1))
		require.NoError(t, err1)
		assert.Equal(t, float64(86), res1)
		assert.Equal(t, ReasonRuleMatch, detail1.Reason)
		res1, detail1, err1 = client.DoubleVariation("ff-test-number", testUser4, float64(-1))
		require.NoError(t, err1)
		assert.Equal(t, float64(9999), res1)
		assert.Equal(t, ReasonFallthrough, detail1.Reason)
	})
	t.Run("string variation", func(t *testing.T) {
		res, detail, err := client.Variation("ff-test-string", testUser5, "error")
		require.NoError(t, err)
		assert.Equal(t, "phone number", res)
		assert.Equal(t, ReasonRuleMatch, detail.Reason)
		res, detail, err = client.Variation("ff-test-string", testUser6, "error")
		require.NoError(t, err)
		assert.Equal(t, "phone number", res)
		assert.Equal(t, ReasonRuleMatch, detail.Reason)
		res, detail, err = client.Variation("ff-test-string", testUser7, "error")
		require.NoError(t, err)
		assert.Equal(t, "email", res)
		assert.Equal(t, ReasonRuleMatch, detail.Reason)
		res, detail, err = client.Variation("ff-test-string", testUser1, "error")
		require.NoError(t, err)
		assert.Equal(t, "others", res)
		assert.Equal(t, ReasonFallthrough, detail.Reason)
	})
	t.Run("segment", func(t *testing.T) {
		res, detail, err := client.Variation("ff-test-seg", testUser1, "error")
		require.NoError(t, err)
		assert.Equal(t, "teamA", res)
		assert.Equal(t, ReasonRuleMatch, detail.Reason)
		res, detail, err = client.Variation("ff-test-seg", testUser2, "error")
		require.NoError(t, err)
		assert.Equal(t, "teamB", res)
		assert.Equal(t, ReasonFallthrough, detail.Reason)
		res, detail, err = client.Variation("ff-test-seg", testUser3, "error")
		require.NoError(t, err)
		assert.Equal(t, "teamA", res)
		assert.Equal(t, ReasonRuleMatch, detail.Reason)
		res, detail, err = client.Variation("ff-test-seg", testUser4, "error")
		require.NoError(t, err)
		assert.Equal(t, "teamB", res)
		assert.Equal(t, ReasonFallthrough, detail.Reason)
	})
	t.Run("json variation", func(t *testing.T) {
		dummy := Dummy{}
		res, detail, err := client.JsonVariation("ff-test-json", testUser1, dummy)
		require.NoError(t, err)
		code := res.(Dummy).Code
		assert.Equal(t, 200, code)
		assert.Equal(t, ReasonFallthrough, detail.Reason)
		res, detail, err = client.JsonVariation("ff-test-json", testUser2, dummy)
		require.NoError(t, err)
		code = res.(Dummy).Code
		assert.Equal(t, 404, code)
		assert.Equal(t, ReasonFallthrough, detail.Reason)
	})
	t.Run("check if flag is known", func(t *testing.T) {
		assert.True(t, client.IsFlagKnown("ff-test-bool"))
		assert.True(t, client.IsFlagKnown("ff-test-number"))
		assert.True(t, client.IsFlagKnown("ff-test-string"))
		assert.True(t, client.IsFlagKnown("ff-test-json"))
		assert.True(t, client.IsFlagKnown("ff-test-seg"))
		assert.False(t, client.IsFlagKnown("ff-not-existed"))
	})
	t.Run("all latest flag values for a given user", func(t *testing.T) {
		allState, err := client.AllLatestFlagsVariations(testUser1)
		require.NoError(t, err)
		assert.True(t, allState.IsSuccess())
		assert.Equal(t, "OK", allState.Reason())
		res, detail, _ := allState.GetBoolVariation("ff-test-bool", false)
		assert.True(t, res)
		assert.Equal(t, ReasonTargetMatch, detail.Reason)
		res1, detail1, _ := allState.GetIntVariation("ff-test-number", -1)
		assert.Equal(t, 1, res1)
		assert.Equal(t, ReasonRuleMatch, detail1.Reason)
		res2, detail2, _ := allState.GetDoubleVariation("ff-test-number", float64(-1))
		assert.Equal(t, float64(1), res2)
		assert.Equal(t, ReasonRuleMatch, detail2.Reason)
		res3, detail3, _ := allState.GetStringVariation("ff-test-string", "error")
		assert.Equal(t, "others", res3)
		assert.Equal(t, ReasonFallthrough, detail3.Reason)
		res4, detail4, _ := allState.GetJsonVariation("ff-test-json", Dummy{})
		code := res4.(Dummy).Code
		assert.Equal(t, 200, code)
		assert.Equal(t, ReasonFallthrough, detail4.Reason)
		res5, detail5, _ := allState.GetStringVariation("ff-test-seg", "error")
		assert.Equal(t, "teamA", res5)
		assert.Equal(t, ReasonRuleMatch, detail5.Reason)
		res6, detail6, err6 := allState.GetStringVariation("ff-not-existed", "error")
		assert.Equal(t, err6, flagNotFound)
		assert.Equal(t, "error", res6)
		assert.Equal(t, ReasonFlagNotFound, detail6.Reason)
	})
	t.Run("argument error", func(t *testing.T) {
		res, detail, _ := client.Variation("ff-not-existed", testUser1, "error")
		assert.Equal(t, "error", res)
		assert.Equal(t, ReasonFlagNotFound, detail.Reason)
		res, detail, _ = client.Variation("ff-test-string", interfaces.FBUser{}, "error")
		assert.Equal(t, "error", res)
		assert.Equal(t, ReasonUserNotSpecified, detail.Reason)
		res1, detail1, _ := client.BoolVariation("ff-test-string", testUser1, false)
		assert.Equal(t, false, res1)
		assert.Equal(t, ReasonWrongType, detail1.Reason)
		allState, _ := client.AllLatestFlagsVariations(interfaces.FBUser{})
		assert.False(t, allState.IsSuccess())
		assert.Equal(t, ReasonUserNotSpecified, allState.Reason())
	})
	_ = client.Close()
}

func TestFBTrackEvent(t *testing.T) {
	parseFlagEvent := func(bytes []byte) []interfaces.Event {
		var events []*insight.FlagEvent
		_ = json.Unmarshal(bytes, &events)
		ret := make([]interfaces.Event, len(events))
		for i, event := range events {
			ret[i] = event
		}
		return ret
	}

	parseUserEvent := func(bytes []byte) []interfaces.Event {
		var events []*insight.UserEvent
		_ = json.Unmarshal(bytes, &events)
		ret := make([]interfaces.Event, len(events))
		for i, event := range events {
			ret[i] = event
		}
		return ret
	}

	parseMetricEvent := func(bytes []byte) []interfaces.Event {
		var events []*insight.MetricEvent
		_ = json.Unmarshal(bytes, &events)
		ret := make([]interfaces.Event, len(events))
		for i, event := range events {
			ret[i] = event
		}
		return ret
	}

	t.Run("send flag event 1", func(t *testing.T) {
		sender := insight2.NewMockSender()
		sender.SetParseJson(parseFlagEvent)
		config := FBConfig{
			StartWait:               200 * time.Millisecond,
			DataStorageFactory:      datastorage.NewMockDataStorageBuilder(),
			DataSynchronizerFactory: datasynchronization.NewMockStreamingBuilder(true, true, 10*time.Millisecond),
			InsightProcessorFactory: insight2.NewMockInsightProcessorFactory(sender, 100, 100*time.Millisecond),
		}
		client, err := MakeCustomFBClient(fakeEnvSecret, "ws://fake-url", "http://fake-url", config)
		require.NoError(t, err)
		assert.True(t, client.IsInitialized())
		res, _, _ := client.Variation("ff-test-string", testUser1, "error")
		assert.Equal(t, "others", res)
		info, ok := sender.GetLatestSendingInfo(200 * time.Millisecond)
		assert.True(t, ok)
		assert.Equal(t, 1, info.Size())
		assert.True(t, info.Contains("test-user-1"))
		_ = client.Close()
	})
	t.Run("send flag event 2", func(t *testing.T) {
		sender := insight2.NewMockSender()
		sender.SetParseJson(parseFlagEvent)
		config := FBConfig{
			StartWait:               200 * time.Millisecond,
			DataStorageFactory:      datastorage.NewMockDataStorageBuilder(),
			DataSynchronizerFactory: datasynchronization.NewMockStreamingBuilder(true, true, 10*time.Millisecond),
			InsightProcessorFactory: insight2.NewMockInsightProcessorFactory(sender, 100, 100*time.Millisecond),
		}
		client, err := MakeCustomFBClient(fakeEnvSecret, "ws://fake-url", "http://fake-url", config)
		require.NoError(t, err)
		assert.True(t, client.IsInitialized())
		allState, _ := client.AllLatestFlagsVariations(testUser1)
		assert.True(t, allState.IsSuccess())
		res, _, _ := allState.GetStringVariation("ff-test-string", "error")
		assert.Equal(t, "others", res)
		info, ok := sender.GetLatestSendingInfo(200 * time.Millisecond)
		assert.True(t, ok)
		assert.Equal(t, 1, info.Size())
		assert.True(t, info.Contains("test-user-1"))
		_ = client.Close()
	})
	t.Run("send user event", func(t *testing.T) {
		sender := insight2.NewMockSender()
		sender.SetParseJson(parseUserEvent)
		config := FBConfig{
			StartWait:               200 * time.Millisecond,
			DataStorageFactory:      datastorage.NewMockDataStorageBuilder(),
			DataSynchronizerFactory: datasynchronization.NewMockStreamingBuilder(true, true, 10*time.Millisecond),
			InsightProcessorFactory: insight2.NewMockInsightProcessorFactory(sender, 100, 100*time.Millisecond),
		}
		client, err := MakeCustomFBClient(fakeEnvSecret, "ws://fake-url", "http://fake-url", config)
		require.NoError(t, err)
		assert.True(t, client.IsInitialized())
		_ = client.Identify(testUser1)
		info, ok := sender.GetLatestSendingInfo(200 * time.Millisecond)
		assert.True(t, ok)
		assert.Equal(t, 1, info.Size())
		assert.True(t, info.Contains("test-user-1"))
		_ = client.Close()
	})

	t.Run("send metric event", func(t *testing.T) {
		sender := insight2.NewMockSender()
		sender.SetParseJson(parseMetricEvent)
		config := FBConfig{
			StartWait:               200 * time.Millisecond,
			DataStorageFactory:      datastorage.NewMockDataStorageBuilder(),
			DataSynchronizerFactory: datasynchronization.NewMockStreamingBuilder(true, true, 10*time.Millisecond),
			InsightProcessorFactory: insight2.NewMockInsightProcessorFactory(sender, 100, 100*time.Millisecond),
		}
		client, err := MakeCustomFBClient(fakeEnvSecret, "ws://fake-url", "http://fake-url", config)
		require.NoError(t, err)
		assert.True(t, client.IsInitialized())
		_ = client.TrackNumericMetric(testUser1, "metric 1", 1.0)
		_ = client.Flush()
		time.Sleep(10 * time.Millisecond)
		_ = client.TrackNumericMetrics(testUser2, map[string]float64{"metric 2": 1.0, "metric 3": 1.0})
		_ = client.Flush()
		info, _ := sender.GetLatestSendingInfo(200 * time.Millisecond)
		assert.Equal(t, 1, info.Size())
		assert.True(t, info.Contains("test-user-1"))
		info, _ = sender.GetLatestSendingInfo(200 * time.Millisecond)
		assert.Equal(t, 1, info.Size())
		assert.True(t, info.Contains("test-user-2"))
		_ = client.Close()
	})
}
