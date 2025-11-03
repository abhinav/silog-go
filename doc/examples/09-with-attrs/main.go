// # With Attributes
//
// The With method is used to pre-attach attributes
// for shared context across multiple log statements.
package main

import (
	"log/slog"
	"os"

	"go.abhg.dev/log/silog"
)

func main() {
	h := silog.NewHandler(os.Stderr, &silog.HandlerOptions{
		Style: silog.DefaultStyle(nil),
		Level: slog.LevelDebug,
	})
	log := slog.New(h)

	// <EXAMPLE>
	requestLog := log.With("request_id", "abc-123", "user", "alice@example.com", "session", "xyz-789")

	requestLog.Info("Request started")
	requestLog.Info("Processing payment")
	requestLog.Info("Request completed", "duration", "245ms")
	// </EXAMPLE>
}
