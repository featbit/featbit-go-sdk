package interfaces

import (
	"context"
	"log/slog"
)

// Logger receives SDK log records. A *slog.Logger implements this interface.
// Implementations control level filtering, formatting, and output, and must be
// safe for concurrent calls. The SDK does not close application-owned loggers.
type Logger interface {
	// Log receives a non-nil context, a level, a message, and structured fields
	// as alternating string keys and values. The SDK uses the standard slog
	// levels, plus slog.Level(-8) for trace messages.
	Log(ctx context.Context, level slog.Level, msg string, args ...any)
}

var _ Logger = (*slog.Logger)(nil)

// LoggerProvider is an optional interface implemented by the SDK's component
// context. Custom factories can use it to obtain their client's logger without
// requiring other Context implementations to provide one.
type LoggerProvider interface {
	GetLogger() Logger
}
