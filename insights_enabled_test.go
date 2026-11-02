package featbit

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/featbit/featbit-go-sdk/v2/factories"
	"github.com/featbit/featbit-go-sdk/v2/fixtures"
	"github.com/featbit/featbit-go-sdk/v2/interfaces"
	"github.com/featbit/featbit-go-sdk/v2/internal/types/data"
	"github.com/featbit/featbit-go-sdk/v2/internal/types/insight"
	"github.com/stretchr/testify/require"
)

type insightsRecorder struct{ events []interfaces.Event }

func (r *insightsRecorder) Send(e interfaces.Event) { r.events = append(r.events, e) }
func (r *insightsRecorder) Flush()                  {}
func (r *insightsRecorder) Close() error            { return nil }
func (r *insightsRecorder) CreateInsightProcessor(interfaces.Context) (interfaces.InsightProcessor, error) {
	return r, nil
}

func insightsClient(t *testing.T) (*FBClient, *insightsRecorder) {
	t.Helper()
	r := &insightsRecorder{}
	c, err := MakeCustomFBClient(fakeEnvSecret, "ws://fake", "http://fake", FBConfig{
		DataSynchronizerFactory: factories.ExternalDataSynchronization(),
		InsightProcessorFactory: r,
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, c.Close()) })
	return c, r
}

func insightsData(t *testing.T, setting string, version int64) data.Data {
	t.Helper()
	b, err := fixtures.LoadFBClientTestData()
	require.NoError(t, err)
	var envelope map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &envelope))
	d := envelope["data"].(map[string]interface{})
	for _, raw := range d["featureFlags"].([]interface{}) {
		f := raw.(map[string]interface{})
		delete(f, "insightsEnabled")
		f["updatedAt"] = time.Unix(version, 0).UTC().Format(time.RFC3339)
		if setting != "missing" {
			var value interface{}
			require.NoError(t, json.Unmarshal([]byte(setting), &value))
			f["insightsEnabled"] = value
		}
	}
	b, err = json.Marshal(d)
	require.NoError(t, err)
	var result data.Data
	require.NoError(t, json.Unmarshal(b, &result))
	return result
}

func TestInsightsEvaluationEntries(t *testing.T) {
	for _, setting := range []string{"missing", "true", "false", "null", "0", `"false"`} {
		t.Run(setting, func(t *testing.T) {
			c, r := insightsClient(t)
			d := insightsData(t, setting, 1)
			require.True(t, c.dataUpdater.Init(d.ToStorageType(), 1))
			_, _, err := c.Variation("ff-test-string", testUser1, "fallback")
			require.NoError(t, err)
			_, _, err = c.BoolVariation("ff-test-bool", testUser1, false)
			require.NoError(t, err)
			_, _, err = c.IntVariation("ff-test-number", testUser1, 0)
			require.NoError(t, err)
			_, _, err = c.DoubleVariation("ff-test-number", testUser1, 0)
			require.NoError(t, err)
			_, _, err = c.JsonVariation("ff-test-json", testUser1, Dummy{})
			require.NoError(t, err)
			states, err := c.AllLatestFlagsVariations(testUser1)
			require.NoError(t, err)
			count := 5
			if setting == "false" {
				count = 0
			}
			require.Len(t, r.events, count)
			_, _, err = states.GetBoolVariation("ff-test-bool", false)
			require.NoError(t, err)
			if setting != "false" {
				count++
			}
			require.Len(t, r.events, count)
			require.NoError(t, c.TrackNumericMetric(testUser1, "purchase", 0))
			require.NoError(t, c.Identify(testUser1))
			require.Len(t, r.events, count+2)
		})
	}
}

func TestInsightsUpdatesPreserveSnapshot(t *testing.T) {
	c, r := insightsClient(t)
	d := insightsData(t, "true", 1)
	require.True(t, c.dataUpdater.Init(d.ToStorageType(), 1))
	enabled, err := c.AllLatestFlagsVariations(testUser1)
	require.NoError(t, err)
	d = insightsData(t, "false", 2)
	require.True(t, c.dataUpdater.Upsert(data.Features, "ff-test-bool", d.ToStorageType()[data.Features]["ff-test-bool"], 2))
	disabled, err := c.AllLatestFlagsVariations(testUser1)
	require.NoError(t, err)
	_, _, err = c.BoolVariation("ff-test-bool", testUser1, false)
	require.NoError(t, err)
	require.Empty(t, r.events)
	_, _, err = enabled.GetBoolVariation("ff-test-bool", false)
	require.NoError(t, err)
	require.Len(t, r.events, 1)
	d = insightsData(t, "true", 3)
	require.True(t, c.dataUpdater.Upsert(data.Features, "ff-test-bool", d.ToStorageType()[data.Features]["ff-test-bool"], 3))
	_, _, err = disabled.GetBoolVariation("ff-test-bool", false)
	require.NoError(t, err)
	require.Len(t, r.events, 1)
	_, _, err = c.BoolVariation("ff-test-bool", testUser1, false)
	require.NoError(t, err)
	require.Len(t, r.events, 2)
}

func TestInsightsConversionFailure(t *testing.T) {
	c, r := insightsClient(t)
	d := insightsData(t, "true", 1)
	require.True(t, c.dataUpdater.Init(d.ToStorageType(), 1))
	_, _, err := c.JsonVariation("ff-test-string", testUser1, map[string]interface{}{})
	require.Error(t, err)
	require.Empty(t, r.events)
	states, err := c.AllLatestFlagsVariations(testUser1)
	require.NoError(t, err)
	_, _, err = states.GetJsonVariation("ff-test-string", map[string]interface{}{})
	require.Error(t, err)
	require.Empty(t, r.events)
}

func TestInsightsPreserveEvaluationPaths(t *testing.T) {
	for _, tc := range []struct {
		flag *data.FeatureFlag
		user interfaces.FBUser
	}{
		{disabledFlag, user1}, {flag, user2}, {flag, user3}, {flag, user10},
	} {
		copy := *tc.flag
		enabledEvent := insight.NewFlagEvent(insight.ConvertFBUserToEventUser(&tc.user))
		before := eval.evaluate(&copy, &tc.user, enabledEvent)
		copy.InsightsEnabled = false
		disabledEvent := insight.NewFlagEvent(insight.ConvertFBUserToEventUser(&tc.user))
		after := eval.evaluate(&copy, &tc.user, disabledEvent)
		require.Equal(t, before, after)
		require.True(t, enabledEvent.IsSendEvent())
		require.False(t, disabledEvent.IsSendEvent())
	}
}

func TestInsightsDecodeResetsDefault(t *testing.T) {
	var flag data.FeatureFlag
	require.NoError(t, json.Unmarshal([]byte(`{}`), &flag))
	require.True(t, flag.InsightsEnabled)
	for _, setting := range []string{"true", "null", "0", `"false"`, "missing"} {
		require.NoError(t, json.Unmarshal([]byte(`{"insightsEnabled":false}`), &flag))
		require.False(t, flag.InsightsEnabled)
		payload := `{}`
		if setting != "missing" {
			payload = `{"insightsEnabled":` + setting + `}`
		}
		require.NoError(t, json.Unmarshal([]byte(payload), &flag))
		require.True(t, flag.InsightsEnabled)
	}
}

func TestInsightsExternalJSONOffline(t *testing.T) {
	for _, setting := range []string{"missing", "true", "false"} {
		t.Run(setting, func(t *testing.T) {
			r := &insightsRecorder{}
			c, err := MakeCustomFBClient(fakeEnvSecret, "ws://fake", "http://fake", FBConfig{Offline: true, InsightProcessorFactory: r})
			require.NoError(t, err)
			defer c.Close()
			d := insightsData(t, setting, 1)
			payload, err := json.Marshal(data.All{Message: data.Message{MessageType: "data-sync"}, Data: d})
			require.NoError(t, err)
			var raw map[string]interface{}
			require.NoError(t, json.Unmarshal(payload, &raw))
			for _, flag := range raw["data"].(map[string]interface{})["featureFlags"].([]interface{}) {
				flag.(map[string]interface{})["updatedAt"] = "2026-11-02T00:00:00Z"
			}
			payload, err = json.Marshal(raw)
			require.NoError(t, err)
			ok, err := c.InitializeFromExternalJson(string(payload))
			require.NoError(t, err)
			require.True(t, ok)
			require.Equal(t, setting != "false", c.getFlag("ff-test-bool").InsightsEnabled)
			_, _, err = c.BoolVariation("ff-test-bool", testUser1, false)
			require.NoError(t, err)
			require.Empty(t, r.events)
		})
	}
}
