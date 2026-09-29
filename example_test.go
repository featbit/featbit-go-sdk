package featbit_test

import (
	"log/slog"
	"os"

	featbit "github.com/featbit/featbit-go-sdk/v2"
)

func ExampleMakeCustomFBClient_logger() {
	config := *featbit.DefaultFBConfig
	config.Offline = true // This example does not connect to a FeatBit server.
	config.Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})).With("component", "featbit")

	// The injected handler controls the log level, format, and destination.
	client, err := featbit.MakeCustomFBClient("", "", "", config)
	if err != nil {
		panic(err)
	}
	defer client.Close()
}
