// Package silog provides a slog.Handler implementation
// that produces human-readable, logfmt-style output.
//
// Its features include:
//
//   - Colored output
//   - Custom log levels
//   - Multi-line messages and attributes
//   - Prefixes for log messages
//
// # Usage
//
// Construct a silog.Handler with [NewHandler]:
//
//	handler := silog.NewHandler(os.Stderr, silog.Options{
//		Level:  silog.LevelInfo,
//	})
//
// Then use it with a slog.Logger:
//
//	logger := slog.New(handler)
package silog
