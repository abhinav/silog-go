# Examples

## Basic Levels

Built-in levels are demonstrated with their default color-coded styling.


```go
logger.Debug("Starting background task")
logger.Info("Server listening on :8080")
logger.Warn("Connection pool nearing capacity")
logger.Error("Failed to process request")
```

![screenshot of 01-basic-levels](img/examples/01-basic-levels.svg)

## Multiline Errors

Multi-line error messages are formatted with each line receiving the timestamp and level prefix.


```go
logger.Error("Database connection failed:\nConnection refused\n  at db.Connect()\n  at main.startup()")
```

![screenshot of 02-multiline-errors](img/examples/02-multiline-errors.svg)

## Multiline Attributes

Multi-line attribute values are formatted with pipe-prefixed indentation for readability.


```go
logger.Info("Request completed",
	"sql", "SELECT *\nFROM users\nWHERE active = true",
	"duration", "45ms",
)
```

![screenshot of 03-multiline-attrs](img/examples/03-multiline-attrs.svg)

## Grouped Attributes

Nested grouped attributes are created using WithGroup and slog.Group, rendered with dot-notation for hierarchical attribute organization.


```go
logger := slog.New(handler).WithGroup("request").WithGroup("headers")

logger.Info("Incoming request",
	slog.Group("body",
		"method", "POST",
		"path", "/api/users",
	),
)
```

![screenshot of 04-grouped-attrs](img/examples/04-grouped-attrs.svg)

## Error Highlighting

The "error" attribute is automatically highlighted in red when using the default style.


```go
logger.Info("Operation failed",
	"error", "connection timeout",
	"retry_count", 3,
	"timeout", "30s",
)
```

![screenshot of 05-error-highlight](img/examples/05-error-highlight.svg)

## Prefixes

Handler-level prefixes are used to distinguish log messages posted by different subsystems or components.


```go
dbHandler := baseHandler.SetPrefix("database")
cacheHandler := baseHandler.SetPrefix("cache")

dbLogger := slog.New(dbHandler)
cacheLogger := slog.New(cacheHandler)

dbLogger.Info("Connection pool initialized")
cacheLogger.Warn("Cache miss rate high")
```

![screenshot of 06-prefixes](img/examples/06-prefixes.svg)

## Multiline Prefixes

Prefixes are preserved across multi-line messages, appearing on each line of output.


```go
log := slog.New(h.SetPrefix("worker"))

log.Info("Task completed:\n- Processed 1000 items\n- Generated 50 reports\n- Sent 25 notifications")
```

![screenshot of 07-multiline-prefix](img/examples/07-multiline-prefix.svg)

## Complex Example

A complex real-world example demonstrating combining multi-line error attributes, structured key-value pairs, and error highlighting.


```go
log.Error("API request failed",
	"method", "POST",
	"path", "/api/orders",
	"error", "validation failed:\n  - invalid email\n  - missing required field: address",
	"user_id", "12345",
	"duration", "125ms",
)
```

![screenshot of 08-complex-example](img/examples/08-complex-example.svg)

## With Attributes

The With method is used to pre-attach attributes for shared context across multiple log statements.


```go
requestLog := log.With("request_id", "abc-123", "user", "alice@example.com", "session", "xyz-789")

requestLog.Info("Request started")
requestLog.Info("Processing payment")
requestLog.Info("Request completed", "duration", "245ms")
```

![screenshot of 09-with-attrs](img/examples/09-with-attrs.svg)

## Custom Level

A custom log level can be defined with its own label and message style.


```go
style.LevelLabels[LevelTrace] = renderer.NewStyle().SetString("TRC")
style.Messages[LevelTrace] = style.Messages[slog.LevelDebug]

logger.Log(context.Background(), LevelTrace, "Entering function")
```

![screenshot of 10-custom-level](img/examples/10-custom-level.svg)

## Level Offset

WithLevelOffset dynamically downgrades log levels for testing or temporarily reducing log verbosity.


```go
offsetHandler := baseHandler.WithLevelOffset(-4)
logger := slog.New(offsetHandler)

logger.Error("This appears as WARNING")
logger.Warn("This appears as INFO")
logger.Info("This appears as DEBUG")
```

![screenshot of 11-level-offset](img/examples/11-level-offset.svg)

## Nested Multiline

Multi-line attribute values within deeply nested groups, show how formatting is preserved through multiple levels.


```go
logger := baseLogger.WithGroup("service").WithGroup("database")

logger.Info("Query executed",
	slog.Group("details",
		"query", "SELECT id, name\nFROM users\nORDER BY created_at",
		"rows", 42))
```

![screenshot of 12-nested-multiline](img/examples/12-nested-multiline.svg)

## No Time

ReplaceAttr can suppress the timestamp from log output.


```go
ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
	// Remove the time attribute from log output.
	if len(groups) == 0 && attr.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return attr
},
```

![screenshot of 13-no-time](img/examples/13-no-time.svg)

