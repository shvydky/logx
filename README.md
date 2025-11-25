# logx

`logx` is a lightweight structured logging helper for Go, built on top of the standard `log/slog` API (Go 1.21+).
It enhances the default logging capabilities with hierarchical log-level routing and automatic package/type attribution.

---

## Features

### • Hierarchical log-level routing
`logx` supports overriding log levels at three levels:

1. **Global log level** (default for everything)
2. **Package-level overrides** (for example, `myapp/internal/db`)
3. **Type-level overrides** (for example, `myapp/internal/db.ConnectionPool`)

Priority:

```
type-level > package-level > global-level
```

---

### • Automatic package and type metadata
`logx.For()` enriches log entries with:

```json
{
  "pkg": "myapp/module/Type",
  "type": "Type",
}
```

This eliminates the need for per-module logger initialization.

---

### • Development vs production output
- Set `Config.Pretty` to enable human-readable colorized logs (via `tint`).
- Unset `Config.Pretty` for pure JSON structured logs.

---

## Installation

```bash
go get github.com/your-org/logx
```

---

## Quick Start

### Initialize logger

```go
logx.Init(logx.Config{
    Pretty:      true,
    DefaultLevel: slog.LevelInfo, // global level
    Levels: map[string]slog.Level{
        // Package-level override
        "myapp/worker": slog.LevelDebug,

        // Type-level override
        "myapp/db.ConnectionPool": slog.LevelWarn,
    },
})
```

---

## Struct-based loggers

```go
type Worker struct{}

func (w *Worker) Run() {
    logger := logx.For(w)
    logger.Info("worker started")
}
```

---


## Example output

### Development mode

```
INFO worker.go:42 worker started pkg=myapp/worker type=Worker
```

### Production (JSON)

```json
{
  "time": "2025-01-08T12:00:34.123Z",
  "level": "INFO",
  "msg": "worker started",
  "pkg": "myapp/worker",
  "type": "Worker",
}
```

---

## How it works

`logx` wraps the default `slog.Handler` and applies hierarchical routing:

```
if type-level rule exists:
    use it
else if package-level rule exists:
    use it
else:
    use global level
```

Fast-path evaluation ensures minimal overhead.

---
## Requirements

- Go 1.21+
- `github.com/lmittmann/tint`

---

## License

MIT License

---

## Contributing

Issues and pull requests are welcome.
