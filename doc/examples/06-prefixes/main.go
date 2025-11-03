// # Prefixes
//
// Handler-level prefixes are used to distinguish
// log messages posted by different subsystems or components.
package main

import (
	"log/slog"
	"os"

	"go.abhg.dev/log/silog"
)

func main() {
	baseHandler := silog.NewHandler(os.Stderr, &silog.HandlerOptions{
		Level: slog.LevelDebug,
		Style: silog.DefaultStyle(nil),
	})

	// <EXAMPLE>
	dbHandler := baseHandler.SetPrefix("database")
	cacheHandler := baseHandler.SetPrefix("cache")

	dbLogger := slog.New(dbHandler)
	cacheLogger := slog.New(cacheHandler)

	dbLogger.Info("Connection pool initialized")
	cacheLogger.Warn("Cache miss rate high")
	// </EXAMPLE>
}
