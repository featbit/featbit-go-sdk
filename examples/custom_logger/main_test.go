package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
	"testing"

	featbit "github.com/featbit/featbit-go-sdk/v2"
)

func Example() {
	if err := run(os.Stdout); err != nil {
		panic(err)
	}
	// Output:
	// app: [INFO] FB GO SDK: SDK is in offline mode
	// app: [WARN] FB Go SDK: unknown feature flag; returning default value flag_key=example-flag
	// app: fallback=false error=feature flag not found
	// app: [INFO] FB GO SDK: SDK client is closing
}

func TestLogAdapterFiltersLevels(t *testing.T) {
	var output bytes.Buffer
	adapter := logAdapter{logger: log.New(&output, "app: ", 0), minLevel: slog.LevelWarn}
	for _, level := range []slog.Level{featbit.LevelTrace, slog.LevelDebug, slog.LevelInfo} {
		adapter.Log(context.Background(), level, "hidden")
	}
	adapter.Log(context.Background(), slog.LevelWarn, "visible warning")
	adapter.Log(context.Background(), slog.LevelError, "visible error")
	const want = "app: [WARN] visible warning\napp: [ERROR] visible error\n"
	if got := output.String(); got != want {
		t.Fatalf("filtered logs = %q, want %q", got, want)
	}
}

func TestLogAdapterPreservesTraceAndFields(t *testing.T) {
	var output bytes.Buffer
	adapter := logAdapter{logger: log.New(&output, "app: ", 0), minLevel: featbit.LevelTrace}
	adapter.Log(context.Background(), featbit.LevelTrace, "diagnostic",
		"flag_key", "example-flag", "enabled", false, "attempt", 2, "error", errors.New("offline"))
	const want = "app: [TRACE] diagnostic flag_key=example-flag enabled=false attempt=2 error=offline\n"
	if got := output.String(); got != want {
		t.Fatalf("trace log = %q, want %q", got, want)
	}
}

func TestLogAdapterConcurrentCalls(t *testing.T) {
	var output bytes.Buffer
	adapter := logAdapter{logger: log.New(&output, "app: ", 0), minLevel: slog.LevelInfo}
	const workers = 20
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			adapter.Log(context.Background(), slog.LevelInfo, "concurrent", "worker", worker)
		}(i)
	}
	wg.Wait()
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != workers {
		t.Fatalf("got %d log lines, want %d", len(lines), workers)
	}
	counts := make(map[string]int)
	for _, line := range lines {
		counts[line]++
	}
	for i := 0; i < workers; i++ {
		line := fmt.Sprintf("app: [INFO] concurrent worker=%d", i)
		if counts[line] != 1 {
			t.Errorf("log line %q occurred %d times, want once", line, counts[line])
		}
	}
}
