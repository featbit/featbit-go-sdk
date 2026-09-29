package factories_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/featbit/featbit-go-sdk/v2/factories"
	"github.com/featbit/featbit-go-sdk/v2/interfaces"
	"github.com/featbit/featbit-go-sdk/v2/internal/types/insight"
	"github.com/gorilla/websocket"
)

type componentLogBuffer struct {
	mu sync.Mutex
	bytes.Buffer
}

func (b *componentLogBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Buffer.Write(p)
}

func (b *componentLogBuffer) snapshot() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Buffer.String()
}

type localEventTransport struct {
	sent chan struct{}
}

func (t localEventTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	t.sent <- struct{}{}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("{}")),
		Header:     make(http.Header),
		Request:    request,
	}, nil
}

type componentTestNetwork struct {
	client *http.Client
}

func (n componentTestNetwork) GetHeaders(map[string]string) http.Header { return make(http.Header) }
func (n componentTestNetwork) GetHTTPClient() interfaces.NetworkClient  { return n.client }
func (n componentTestNetwork) GetWebsocketClient() interfaces.NetworkClient {
	return websocket.DefaultDialer
}

// This context intentionally implements only the original Context contract.
type legacyComponentContext struct {
	network interfaces.Network
}

func (c legacyComponentContext) GetEnvSecret() string           { return "test-secret" }
func (c legacyComponentContext) GetStreamingUri() string        { return "ws://example.invalid" }
func (c legacyComponentContext) GetEventUri() string            { return "http://example.invalid" }
func (c legacyComponentContext) GetNetwork() interfaces.Network { return c.network }

type loggingComponentContext struct {
	interfaces.Context
	logger interfaces.Logger
}

func (c loggingComponentContext) GetLogger() interfaces.Logger { return c.logger }

type testComponents struct {
	stream    interfaces.DataSynchronizer
	processor interfaces.InsightProcessor
	sent      <-chan struct{}
}

func newTestComponents(t *testing.T, logger interfaces.Logger) testComponents {
	t.Helper()
	sent := make(chan struct{}, 1)
	var ctx interfaces.Context = legacyComponentContext{network: componentTestNetwork{
		client: &http.Client{Transport: localEventTransport{sent: sent}},
	}}
	if logger != nil {
		ctx = loggingComponentContext{Context: ctx, logger: logger}
	} else if _, ok := ctx.(interfaces.LoggerProvider); ok {
		t.Fatal("legacy test context unexpectedly implements LoggerProvider")
	}
	stream, err := factories.NewStreamingBuilder().CreateDataSynchronizer(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	processor, err := factories.NewInsightProcessorBuilder().FlushInterval(10 * time.Millisecond).CreateInsightProcessor(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stream.Close() })
	t.Cleanup(func() { _ = processor.Close() })
	return testComponents{stream: stream, processor: processor, sent: sent}
}

func exerciseComponents(t *testing.T, components testComponents) {
	t.Helper()
	user, err := interfaces.NewUserBuilder("logger-test-user").Build()
	if err != nil {
		t.Error(err)
		return
	}
	components.processor.Send(insight.NewUserEvent(insight.ConvertFBUserToEventUser(&user)))
	components.processor.Flush()
	select {
	case <-components.sent:
	case <-time.After(3 * time.Second):
		t.Error("background event processor did not send its event")
	}
	_ = components.processor.Close()
	// Closing an unstarted stream exercises logging without any WebSocket I/O.
	_ = components.stream.Close()
}

func TestBuiltInFactoriesKeepConcurrentComponentLoggersSeparate(t *testing.T) {
	var firstLogs, secondLogs componentLogBuffer
	options := &slog.HandlerOptions{Level: slog.LevelDebug}
	first := newTestComponents(t, slog.New(slog.NewJSONHandler(&firstLogs, options)).With("client", "first"))
	second := newTestComponents(t, slog.New(slog.NewJSONHandler(&secondLogs, options)).With("client", "second"))

	var workers sync.WaitGroup
	for _, components := range []testComponents{first, second} {
		workers.Add(1)
		go func() {
			defer workers.Done()
			exerciseComponents(t, components)
		}()
	}
	workers.Wait()

	for name, output := range map[string]string{"first": firstLogs.snapshot(), "second": secondLogs.snapshot()} {
		messages := make(map[string]bool)
		decoder := json.NewDecoder(strings.NewReader(output))
		for decoder.More() {
			var record map[string]any
			if err := decoder.Decode(&record); err != nil {
				t.Fatalf("invalid %s log record: %v", name, err)
			}
			if record["client"] != name {
				t.Errorf("%s received another client's record: %v", name, record)
			}
			messages[record["msg"].(string)] = true
		}
		for _, message := range []string{
			"FB GO SDK: streaming is stopping",
			"event dispatcher is working",
			"sending event payload",
			"sending events ok",
			"FB GO SDK: insight processor is stopping",
		} {
			if !messages[message] {
				t.Errorf("%s application logger did not receive %q", name, message)
			}
		}
	}
}

func TestBuiltInFactoriesAcceptContextWithoutLoggerProvider(t *testing.T) {
	exerciseComponents(t, newTestComponents(t, nil))
}

type componentLoggerFunc func(context.Context, slog.Level, string, ...any)

func (f componentLoggerFunc) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	f(ctx, level, msg, args...)
}

func TestBuiltInFactoriesAcceptLoggerInterface(t *testing.T) {
	var mu sync.Mutex
	messages := make(map[string]slog.Level)
	logger := componentLoggerFunc(func(ctx context.Context, level slog.Level, msg string, args ...any) {
		if ctx == nil {
			t.Error("background component supplied a nil logging context")
		}
		mu.Lock()
		defer mu.Unlock()
		messages[msg] = level
	})
	exerciseComponents(t, newTestComponents(t, logger))
	mu.Lock()
	defer mu.Unlock()
	for message, wantLevel := range map[string]slog.Level{
		"FB GO SDK: streaming is stopping":         slog.LevelInfo,
		"event dispatcher is working":              slog.LevelDebug,
		"sending event payload":                    slog.LevelDebug,
		"sending events ok":                        slog.LevelDebug,
		"FB GO SDK: insight processor is stopping": slog.LevelInfo,
	} {
		if level, ok := messages[message]; !ok || level != wantLevel {
			t.Errorf("custom logger received %q at level %v (present=%v), want %v", message, level, ok, wantLevel)
		}
	}
}
