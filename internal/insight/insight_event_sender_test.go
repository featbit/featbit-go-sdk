package insight

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"testing"
)

func TestEventSenderUsesItsOwnStructuredLogger(t *testing.T) {
	var firstLogs, secondLogs bytes.Buffer
	first := NewEventSenderImp(http.DefaultClient, nil, 0, 0, slog.New(slog.NewJSONHandler(&firstLogs, nil)))
	second := NewEventSenderImp(http.DefaultClient, nil, 0, 0, slog.New(slog.NewJSONHandler(&secondLogs, nil)))
	if _, err := first.PostJson(":", []byte("{}")); err == nil {
		t.Fatal("expected malformed URL error")
	}
	if secondLogs.Len() != 0 {
		t.Fatal("first sender wrote to second sender's logger")
	}
	var record map[string]any
	if err := json.Unmarshal(firstLogs.Bytes(), &record); err != nil {
		t.Fatalf("invalid structured log: %v", err)
	}
	if record["level"] != "ERROR" || record["error"] == nil {
		t.Fatalf("expected ERROR record with error attribute, got %v", record)
	}
	firstSize := firstLogs.Len()
	if _, err := second.PostJson(":", []byte("{}")); err == nil {
		t.Fatal("expected malformed URL error")
	}
	if firstLogs.Len() != firstSize || secondLogs.Len() == 0 {
		t.Fatal("sender logs were not isolated")
	}
}

func TestEventSenderWithoutLogger(t *testing.T) {
	for _, sender := range []*EventSenderImp{
		NewEventSenderImp(http.DefaultClient, nil, 0, 0),
		NewEventSenderImp(http.DefaultClient, nil, 0, 0, nil),
	} {
		if _, err := sender.PostJson(":", []byte("{}")); err == nil {
			t.Fatal("expected malformed URL error")
		}
	}
}
