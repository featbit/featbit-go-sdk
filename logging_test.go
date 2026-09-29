package featbit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type interfaceLogRecord struct {
	ctx     context.Context
	level   slog.Level
	message string
	args    []any
}

// This implementation has no slog.Logger or slog.Handler. Its only logging
// method is the public Logger contract's Log method.
type interfaceTestLogger struct {
	mu      sync.Mutex
	records []interfaceLogRecord
}

func (l *interfaceTestLogger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.records = append(l.records, interfaceLogRecord{ctx, level, msg, append([]any(nil), args...)})
}

func (l *interfaceTestLogger) snapshot() []interfaceLogRecord {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]interfaceLogRecord(nil), l.records...)
}

func TestClientAcceptsLoggerInterface(t *testing.T) {
	logger := &interfaceTestLogger{}
	// A custom implementation owns filtering, including with legacy ERROR set.
	client := newLoggingTestClient(t, FBConfig{Logger: logger, LogLevel: ERROR})
	initializeLoggingTestData(t, client)
	if _, _, err := client.BoolVariation("missing", testUser1, false); err == nil {
		t.Fatal("missing flag should return an error")
	}
	if _, _, err := client.JsonVariation("malformed-json", testUser1, map[string]interface{}{}); err == nil {
		t.Fatal("malformed JSON variation should return an error")
	}

	levels := make(map[slog.Level]bool)
	var sawFlagKey, sawError bool
	for _, record := range logger.snapshot() {
		if record.ctx == nil {
			t.Error("SDK supplied a nil logging context")
		}
		levels[record.level] = true
		if len(record.args)%2 != 0 {
			t.Errorf("SDK fields are not key-value pairs: %v", record.args)
			continue
		}
		for i := 0; i < len(record.args); i += 2 {
			key, ok := record.args[i].(string)
			if !ok {
				t.Errorf("SDK field key is not a string: %v", record.args[i])
			}
			if record.level == slog.LevelWarn && key == "flag_key" && record.args[i+1] == "missing" {
				sawFlagKey = true
			}
			if key == "error" {
				if err, ok := record.args[i+1].(error); ok && err != nil {
					sawError = true
				}
			}
		}
	}
	for _, level := range []slog.Level{slog.LevelInfo, slog.LevelWarn, slog.LevelError} {
		if !levels[level] {
			t.Errorf("custom logger did not receive %s records", level)
		}
	}
	if !sawFlagKey || !sawError {
		t.Errorf("custom logger lost structured fields: flag_key=%v, error=%v", sawFlagKey, sawError)
	}
}

func TestClientLoggerInterfaceConcurrentCalls(t *testing.T) {
	logger := &interfaceTestLogger{}
	client := newLoggingTestClient(t, FBConfig{Logger: logger})
	initializeLoggingTestData(t, client)
	const workers = 8
	const evaluationsPerWorker = 10
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < evaluationsPerWorker; j++ {
				_, _, _ = client.BoolVariation("missing", testUser1, false)
			}
		}()
	}
	wg.Wait()
	warnings := 0
	for _, record := range logger.snapshot() {
		if record.level == slog.LevelWarn {
			warnings++
		}
	}
	if want := workers * evaluationsPerWorker; warnings != want {
		t.Errorf("shared client logged %d warnings, want %d", warnings, want)
	}
}

const loggingTestData = `{
	"messageType": "data-sync",
	"data": {
		"eventType": "full",
		"featureFlags": [{
			"key": "malformed-json",
			"name": "Malformed JSON",
			"variationType": "json",
			"isEnabled": false,
			"disabledVariationId": "bad",
			"variations": [{"id": "bad", "value": "{"}],
			"updatedAt": "2026-01-01T00:00:00Z"
		}],
		"segments": []
	}
}`

// JSONHandler and the buffer both support concurrent writers. Keeping the buffer
// synchronized also lets tests safely inspect records while other clients log.
type loggingTestBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (b *loggingTestBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(p)
}

func (b *loggingTestBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.String()
}

type loggingTestRecord struct {
	Level    string `json:"level"`
	Message  string `json:"msg"`
	ClientID string `json:"client_id"`
}

func (b *loggingTestBuffer) records(t *testing.T) []loggingTestRecord {
	t.Helper()
	decoder := json.NewDecoder(strings.NewReader(b.String()))
	var records []loggingTestRecord
	for {
		var record loggingTestRecord
		if err := decoder.Decode(&record); err == io.EOF {
			return records
		} else if err != nil {
			t.Fatalf("decode application JSON log: %v", err)
		}
		records = append(records, record)
	}
}

func newLoggingTestLogger(clientID string, level slog.Leveler) (*slog.Logger, *loggingTestBuffer) {
	buffer := &loggingTestBuffer{}
	logger := slog.New(slog.NewJSONHandler(buffer, &slog.HandlerOptions{Level: level}))
	return logger.With("client_id", clientID), buffer
}

func newLoggingTestClient(t *testing.T, config FBConfig) *FBClient {
	t.Helper()
	config.Offline = true
	config.StartWait = time.Millisecond
	client, err := MakeCustomFBClient("", "", "", config)
	if err != nil {
		t.Fatalf("create offline client: %v", err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close offline client: %v", err)
		}
	})
	return client
}

func initializeLoggingTestData(t *testing.T, client *FBClient) {
	t.Helper()
	if ok, err := client.InitializeFromExternalJson(loggingTestData); err != nil || !ok {
		t.Fatalf("initialize offline flags: initialized=%v, error=%v", ok, err)
	}
}

func TestClientCustomLoggerReceivesSDKRecords(t *testing.T) {
	applicationDefault := slog.Default()
	logger, buffer := newLoggingTestLogger("application", slog.LevelInfo)
	// The injected handler owns filtering, regardless of the legacy setting.
	client := newLoggingTestClient(t, FBConfig{Logger: logger, LogLevel: ERROR})
	if _, _, err := client.BoolVariation("missing", testUser1, false); err == nil {
		t.Fatal("missing flag should return an error")
	}
	initializeLoggingTestData(t, client)
	if _, _, err := client.JsonVariation("malformed-json", testUser1, map[string]interface{}{}); err == nil {
		t.Fatal("malformed JSON variation should return an error")
	}
	levels := make(map[string]bool)
	for _, record := range buffer.records(t) {
		if record.ClientID != "application" {
			t.Errorf("SDK log lost application logger attributes: %+v", record)
		}
		levels[record.Level] = true
	}
	for _, level := range []string{"INFO", "WARN", "ERROR"} {
		if !levels[level] {
			t.Errorf("application handler did not receive %s SDK logs: %s", level, buffer.String())
		}
	}
	if slog.Default() != applicationDefault {
		t.Error("creating a client replaced the application's global slog logger")
	}
}

func TestClientCustomLoggerSupportsDynamicFiltering(t *testing.T) {
	var level slog.LevelVar
	level.Set(slog.LevelError)
	logger, buffer := newLoggingTestLogger("dynamic", &level)
	client := newLoggingTestClient(t, FBConfig{Logger: logger, LogLevel: TRACE})
	_, _, _ = client.BoolVariation("hidden-warning", testUser1, false)
	if logs := buffer.String(); logs != "" {
		t.Fatalf("handler ERROR threshold should suppress INFO and WARN: %s", logs)
	}
	level.Set(slog.LevelWarn)
	_, _, _ = client.BoolVariation("visible-warning", testUser1, false)
	records := buffer.records(t)
	if len(records) != 1 || records[0].Level != "WARN" {
		t.Fatalf("updated application threshold should admit one WARN record: %+v", records)
	}
}

func TestClientLoggersStayIsolated(t *testing.T) {
	firstLogger, firstLogs := newLoggingTestLogger("first", slog.LevelInfo)
	first := newLoggingTestClient(t, FBConfig{Logger: firstLogger})
	initializeLoggingTestData(t, first)
	state, err := first.AllLatestFlagsVariations(testUser1)
	if err != nil {
		t.Fatalf("capture flag state: %v", err)
	}
	secondLogger, secondLogs := newLoggingTestLogger("second", slog.LevelWarn)
	second := newLoggingTestClient(t, FBConfig{Logger: secondLogger})
	firstBefore, secondBefore := len(firstLogs.records(t)), len(secondLogs.records(t))
	_, _, _ = first.BoolVariation("first-missing", testUser1, false)
	if _, _, err := state.GetJsonVariation("malformed-json", map[string]interface{}{}); err == nil {
		t.Fatal("captured malformed JSON should return an error")
	}
	if got := len(firstLogs.records(t)) - firstBefore; got != 2 {
		t.Errorf("first client and its flag snapshot should each log to first handler, got %d new records", got)
	}
	if got := len(secondLogs.records(t)); got != secondBefore {
		t.Errorf("first client logs leaked into second handler: %s", secondLogs.String())
	}
	firstBefore = len(firstLogs.records(t))
	_, _, _ = second.BoolVariation("second-missing", testUser1, false)
	if got := len(firstLogs.records(t)); got != firstBefore {
		t.Errorf("second client logs leaked into first handler: %s", firstLogs.String())
	}
	if got := len(secondLogs.records(t)) - secondBefore; got != 1 {
		t.Errorf("second client should log one warning to its handler, got %d", got)
	}
}

func TestClientDefaultLoggerLevel(t *testing.T) {
	for _, tc := range []struct {
		name     string
		level    int
		wantInfo bool
		wantWarn bool
	}{
		{"zero value INFO", INFO, true, true},
		{"WARN", WARN, false, true},
		{"ERROR", ERROR, false, false},
		{"DEBUG", DEBUG, true, true},
		{"TRACE", TRACE, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output := captureLoggingStdout(t, func() {
				client, err := MakeCustomFBClient("", "", "", FBConfig{
					Offline: true, StartWait: time.Millisecond, LogLevel: tc.level,
				})
				if err != nil {
					t.Fatalf("create offline client: %v", err)
				}
				defer func() {
					if err := client.Close(); err != nil {
						t.Errorf("close offline client: %v", err)
					}
				}()
				_, _, _ = client.BoolVariation("missing", testUser1, false)
				initializeLoggingTestData(t, client)
				_, _, err = client.JsonVariation("malformed-json", testUser1, map[string]interface{}{})
				if err == nil {
					t.Error("malformed JSON variation should return an error")
				}
			})
			for level, want := range map[string]bool{"INFO": tc.wantInfo, "WARN": tc.wantWarn, "ERROR": true} {
				if got := strings.Contains(output, "level="+level); got != want {
					t.Errorf("default logger emitted %s=%v, want %v; logs: %s", level, got, want, output)
				}
			}
		})
	}
}

// This test intentionally runs serially because stdout is a process-wide value.
func captureLoggingStdout(t *testing.T, fn func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("capture stdout: %v", err)
	}
	original := os.Stdout
	os.Stdout = writer
	defer func() {
		os.Stdout = original
		_ = writer.Close()
		_ = reader.Close()
	}()
	var output bytes.Buffer
	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(&output, reader)
		done <- err
	}()
	fn()
	os.Stdout = original
	_ = writer.Close()
	if err := <-done; err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}
	return output.String()
}

func TestClientConcurrentLoggerIsolation(t *testing.T) {
	const clientCount = 12
	const warningsPerClient = 20
	var workers sync.WaitGroup
	start := make(chan struct{})
	buffers := make([]*loggingTestBuffer, clientCount)
	for i := 0; i < clientCount; i++ {
		workers.Add(1)
		go func(index int) {
			defer workers.Done()
			clientID := fmt.Sprintf("client-%d", index)
			logger, buffer := newLoggingTestLogger(clientID, slog.LevelWarn)
			buffers[index] = buffer
			<-start
			client, err := MakeCustomFBClient("", "", "", FBConfig{
				Offline: true, StartWait: time.Millisecond, Logger: logger,
			})
			if err != nil {
				t.Errorf("create %s: %v", clientID, err)
				return
			}
			for j := 0; j < warningsPerClient; j++ {
				_, _, _ = client.BoolVariation(clientID+"-missing", testUser1, false)
			}
			if err := client.Close(); err != nil {
				t.Errorf("close %s: %v", clientID, err)
			}
		}(i)
	}
	close(start)
	workers.Wait()
	for i, buffer := range buffers {
		clientID := fmt.Sprintf("client-%d", i)
		records := buffer.records(t)
		if len(records) != warningsPerClient {
			t.Errorf("%s received %d warning records, want %d", clientID, len(records), warningsPerClient)
		}
		for _, record := range records {
			if record.Level != "WARN" || record.ClientID != clientID {
				t.Errorf("unexpected record for %s: %+v", clientID, record)
			}
		}
	}
}
