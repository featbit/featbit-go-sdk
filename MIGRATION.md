# Migrating from v1 to v2

This branch contains unreleased v2 work. The instructions below describe the
planned migration; they do not imply that a `v2.0.0` release is available.

## Go version and imports

Use Go 1.26.0 or later. Update the SDK module and all SDK imports to include `/v2`:

```go
import (
    featbit "github.com/featbit/featbit-go-sdk/v2"
    "github.com/featbit/featbit-go-sdk/v2/interfaces"
)
```

The client constructors and variation methods retain their existing signatures.
If you use keyed `FBConfig` fields and the default logger, logging configuration
can remain unchanged. Positional `FBConfig` literals must account for the new
`Logger` field; switch to keyed fields to avoid that dependency.

## Logging

`FBConfig.Logger` accepts the public `interfaces.Logger` interface:

```go
type Logger interface {
    Log(ctx context.Context, level slog.Level, msg string, args ...any)
}
```

Pass any implementation directly through `config.Logger`. The SDK supplies a
non-nil context, the log level, the message, and structured fields as alternating
string keys and values. Implementations must support concurrent calls and control
their own filtering, formatting, and output. Applications can implement this
interface on their own logger or write an adapter for an existing logging system;
the SDK does not require a `slog.Handler` or provide router-specific adapters.

The interface uses the standard-library `context` and `log/slog` types, but does
not require the slog logger implementation. `*slog.Logger` already satisfies it.

See [Custom logger](README.md#custom-logger) for configuration and the
[runnable adapter example](examples/custom_logger/main.go).

### Using `log/slog`

Pass your application's `*slog.Logger` through `FBConfig.Logger`:

```go
config := *featbit.DefaultFBConfig
config.Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
})).With("component", "featbit")

client, err := featbit.MakeCustomFBClient(envSecret, streamingURL, eventURL, config)
```

This snippet also requires the standard-library `log/slog` and `os` imports.
Each client and its background components use the configured logger. Creating
another client does not replace the first client's logger or level.

- With a non-nil `Logger`, that implementation controls filtering, format, and
  output. `FBConfig.LogLevel` does not override it.
- With a nil `Logger`, the SDK creates a separate `slog.TextHandler` for each
  client, writing to standard output and using `FBConfig.LogLevel`.
- Default log formatting changes to standard slog `key=value` text. Update any
  log parsers that depend on the previous format.
- The SDK does not replace `slog.Default()` or close application-owned loggers or
  handlers when the client closes. Custom loggers must support concurrent use;
  with slog, this requirement also applies to the handler.

Existing `LogLevel` constants keep their numeric values. They are mapped to
slog levels when creating the default handler:

| `FBConfig.LogLevel` | Existing value | slog level |
| --- | ---: | ---: |
| `featbit.TRACE` | -2 | `featbit.LevelTrace` (-8) |
| `featbit.DEBUG` | -1 | `slog.LevelDebug` (-4) |
| `featbit.INFO` | 0 | `slog.LevelInfo` (0) |
| `featbit.WARN` | 1 | `slog.LevelWarn` (4) |
| `featbit.ERROR` | 2 | `slog.LevelError` (8) |

For a custom implementation, handle `featbit.LevelTrace` (-8) to support trace
records. With slog, use it as the handler threshold to include trace logs.
Do not cast the legacy integer constants to `slog.Level`: their
numeric scales differ. The default SDK handler labels trace records `TRACE`.
A plain slog handler renders that level as `DEBUG-4`;
use the handler's `ReplaceAttr` option if your output requires a `TRACE` label.

## Custom factories

Existing factory and `interfaces.Context` method signatures remain unchanged.
Factories can optionally access the client's logger by checking the additional
`interfaces.LoggerProvider` interface:

```go
var logger interfaces.Logger = slog.Default()
if provider, ok := ctx.(interfaces.LoggerProvider); ok {
    if configured := provider.GetLogger(); configured != nil {
        logger = configured
    }
}
```

Your own implementations of `interfaces.Context` do not need to implement this
optional interface to continue satisfying `Context`.
