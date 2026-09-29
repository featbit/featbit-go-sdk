package datasynchronization

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestRetryStrategiesKeepSeparateStateAndLoggers(t *testing.T) {
	var firstLogs, secondLogs bytes.Buffer
	first := NewWithFirstRetryDelay(time.Second, slog.New(slog.NewJSONHandler(&firstLogs, nil)))
	second := NewWithFirstRetryDelay(10*time.Second, slog.New(slog.NewJSONHandler(&secondLogs, nil)))
	first.jitterRatio = 0
	second.jitterRatio = 0
	first.SetGoodRunAtNow()
	second.SetGoodRunAtNow()

	if delay := first.NextDelay(); delay != 500*time.Millisecond {
		t.Fatalf("first strategy delay = %v, want 500ms", delay)
	}
	if secondLogs.Len() != 0 {
		t.Fatal("first strategy wrote to second strategy's logger")
	}
	if delay := second.NextDelay(); delay != 5*time.Second {
		t.Fatalf("second strategy delay = %v, want 5s", delay)
	}
	if delay := first.NextDelay(); delay != time.Second {
		t.Fatalf("first strategy's second delay = %v, want 1s", delay)
	}
	if strings.Count(firstLogs.String(), "backoff before retry") != 2 || strings.Count(secondLogs.String(), "backoff before retry") != 1 {
		t.Fatal("retry logs were not isolated per strategy")
	}
}

func TestNextDelay(t *testing.T) {
	strategy := NewWithFirstRetryDelay(time.Second)
	strategy.SetGoodRunAtNow()
	delay := strategy.NextDelay()
	assert.True(t, delay < time.Second)
	delay = strategy.NextDelay()
	assert.True(t, delay < 2*time.Second)
	delay = strategy.NextDelay()
	assert.True(t, delay < 4*time.Second)
	delay = strategy.NextDelay()
	assert.True(t, delay < 8*time.Second)
	delay = strategy.NextDelay()
	assert.True(t, delay < 16*time.Second)
	delay = strategy.NextDelay()
	assert.True(t, delay < 32*time.Second)
	delay = strategy.NextDelay()
	assert.True(t, delay < 60*time.Second)
}
