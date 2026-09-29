package interfaces

import "log/slog"

// LoggerProvider is an optional interface implemented by the SDK's component
// context. Custom factories can use it to obtain their client's logger without
// requiring other Context implementations to provide one.
type LoggerProvider interface {
	GetLogger() *slog.Logger
}
