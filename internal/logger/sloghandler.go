package logger

import (
	"context"
	"log/slog"
	"slices"
)

// LogHandler adapts the core logger to the slog.Handler interface.
type LogHandler struct {
	logger LoggerInterface
	attrs  []slog.Attr
}

// NewLogHandler creates a new LogHandler instance.
func NewLogHandler(logger LoggerInterface) *LogHandler {
	return &LogHandler{logger: logger}
}

// Enabled determines whether the handler is enabled for the given log level.
func (h *LogHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

// Handle processes a slog.Record and routes it to the core logger.
func (h *LogHandler) Handle(ctx context.Context, record slog.Record) error {
	// Map slog levels to your core logger levels
	var logFunc func(string, ...any) *LogEntry
	switch record.Level {
	case slog.LevelInfo:
		logFunc = h.logger.Info
	case slog.LevelWarn:
		logFunc = h.logger.Warning
	case slog.LevelError:
		logFunc = h.logger.Error
	default:
		logFunc = h.logger.Default
	}

	// Collect attributes as key-value pairs
	numAttrs := record.NumAttrs() + len(h.attrs)
	attrs := make([]any, 0, numAttrs)

	// Add stored attributes
	for _, attr := range h.attrs {
		attrs = append(attrs, attr.Key, attr.Value)
	}

	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr.Key, attr.Value) // Keep attributes as key-value pairs
		return true
	})

	// Log the message with attributes
	if len(attrs) > 0 {
		logFunc(record.Message).WithAttrs(attrs...)
	} else {
		logFunc(record.Message)
	}
	return nil
}

// WithAttrs is a no-op for this handler.
func (h *LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := *h
	newHandler.attrs = slices.Clone(h.attrs)
	newHandler.attrs = append(newHandler.attrs, attrs...)
	return &newHandler
}

// WithGroup creates a new grouped handler.
func (h *LogHandler) WithGroup(name string) slog.Handler {
	return &groupedLogHandler{
		group:    name,
		original: h,
	}
}

// groupedLogHandler handles attributes with grouped keys.
type groupedLogHandler struct {
	group    string
	original slog.Handler
}

func (h *groupedLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.original.Enabled(ctx, level)
}

func (h *groupedLogHandler) Handle(ctx context.Context, record slog.Record) error {
	// Collect grouped attributes
	groupedAttrs := make([]slog.Attr, 0, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool {
		groupedAttrs = append(groupedAttrs, slog.Attr{
			Key:   h.group + "." + attr.Key,
			Value: attr.Value,
		})
		return true
	})

	// Pass only the modified attributes to the original handler
	clonedHandler := h.original.WithAttrs(groupedAttrs)
	return clonedHandler.Handle(ctx, slog.Record{
		Level:   record.Level,
		Message: record.Message,
	})
}

func (h *groupedLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Group attributes before passing them to the original handler
	groupedAttrs := make([]slog.Attr, len(attrs))
	for i, attr := range attrs {
		groupedAttrs[i] = slog.Attr{
			Key:   h.group + "." + attr.Key,
			Value: attr.Value,
		}
	}
	return h.original.WithAttrs(groupedAttrs)
}

func (h *groupedLogHandler) WithGroup(name string) slog.Handler {
	return &groupedLogHandler{
		group:    h.group + "." + name,
		original: h.original,
	}
}
