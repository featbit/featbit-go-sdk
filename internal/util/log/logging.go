package log

import (
	"log/slog"
	"os"

	"github.com/featbit/featbit-go-sdk/v2/interfaces"
)

// LevelTrace is the SDK's diagnostic level below slog.LevelDebug.
const LevelTrace = slog.Level(-8)

// NewDefault creates a separate handler for each client. The integer levels
// retain the v1 FBConfig.LogLevel values; custom loggers use their own handlers.
func NewDefault(level int) *slog.Logger {
	var threshold slog.Level
	switch {
	case level == -2:
		threshold = LevelTrace
	case level < 0:
		threshold = slog.LevelDebug
	case level == 0:
		threshold = slog.LevelInfo
	case level == 1:
		threshold = slog.LevelWarn
	default:
		threshold = slog.LevelError
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: threshold,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if len(groups) == 0 && attr.Key == slog.LevelKey && attr.Value.Any() == LevelTrace {
				return slog.String(slog.LevelKey, "TRACE")
			}
			return attr
		},
	}))
}

// OrDiscard lets standalone internal components operate without a logger.
func OrDiscard(logger *slog.Logger) *slog.Logger {
	if logger == nil {
		return slog.New(slog.DiscardHandler)
	}
	return logger
}

// FromContext supports existing custom Context implementations without adding
// a required method to that public interface.
func FromContext(ctx interfaces.Context) *slog.Logger {
	if provider, ok := ctx.(interfaces.LoggerProvider); ok {
		return OrDiscard(provider.GetLogger())
	}
	return OrDiscard(nil)
}
