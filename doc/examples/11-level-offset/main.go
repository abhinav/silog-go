// # Level Offset
//
// WithLevelOffset dynamically downgrades log levels
// for testing or temporarily reducing log verbosity.
package main

import (
	"log/slog"
	"os"

	"go.abhg.dev/log/silog"
)

func main() {
	baseHandler := silog.NewHandler(os.Stderr, &silog.HandlerOptions{
		Style: silog.DefaultStyle(nil),
		Level: slog.LevelDebug,
	})

	// <EXAMPLE>
	logger := slog.New(baseHandler.WithLevelOffset(-4))

	logger.Error("This appears as WARNING")
	logger.Warn("This appears as INFO")
	logger.Info("This appears as DEBUG")
	// </EXAMPLE>
}
