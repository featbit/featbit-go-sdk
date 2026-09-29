// Run this offline example from the repository root:
//
//	go run ./examples/custom_logger
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"

	featbit "github.com/featbit/featbit-go-sdk/v2"
	"github.com/featbit/featbit-go-sdk/v2/interfaces"
)

// logAdapter belongs to the application, not the SDK. Replace its body with
// calls to your existing logging system. log.Logger supports concurrent writes;
// minLevel is set before use and is not changed while the client is running.
type logAdapter struct {
	logger   *log.Logger
	minLevel slog.Level
}

var _ interfaces.Logger = logAdapter{}

func (a logAdapter) Log(_ context.Context, level slog.Level, msg string, args ...any) {
	if level < a.minLevel {
		return
	}
	name := level.String()
	if level == featbit.LevelTrace {
		name = "TRACE"
	}
	var line strings.Builder
	fmt.Fprintf(&line, "[%s] %s", name, msg)
	// The SDK supplies fields as alternating string keys and values.
	for i := 0; i+1 < len(args); i += 2 {
		fmt.Fprintf(&line, " %s=%v", args[i], args[i+1])
	}
	a.logger.Print(line.String())
}

func run(output io.Writer) error {
	appLogger := log.New(output, "app: ", 0)
	config := *featbit.DefaultFBConfig
	config.Offline = true // No server key or network connection is required.
	config.Logger = logAdapter{logger: appLogger, minLevel: slog.LevelInfo}

	client, err := featbit.MakeCustomFBClient("", "", "", config)
	if client != nil {
		defer client.Close()
	}
	if err != nil {
		return err
	}

	// Initialize an empty environment so evaluating a missing flag emits a
	// warning with a structured flag_key field through our adapter.
	initialized, err := client.InitializeFromExternalJson(`{
		"messageType": "data-sync",
		"data": {"eventType": "full", "featureFlags": [], "segments": []}
	}`)
	if err != nil {
		return err
	}
	if !initialized {
		return fmt.Errorf("offline flag initialization failed")
	}
	user, err := interfaces.NewUserBuilder("example-user").Build()
	if err != nil {
		return err
	}
	value, _, evaluationErr := client.BoolVariation("example-flag", user, false)
	appLogger.Printf("fallback=%t error=%v", value, evaluationErr)
	return nil
}

func main() {
	if err := run(os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
