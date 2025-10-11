// pkg/logger/slog_logger.go
package applogger

import (
	"fmt"
	"log/slog"
	"os"
)

type slogLogger struct {
	logger *slog.Logger
}

// NewSlogLogger คือโรงงานสร้างนักข่าวสายโปรดักชัน
func NewSlogLogger() Logger {
	return &slogLogger{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true, // บอกให้ slog แสดงไฟล์และบรรทัดให้ด้วย
		})),
	}
}

// --- Structured Logging Methods ---
func (l *slogLogger) Debug(msg string, args ...any)   { l.logger.Debug(msg, args...) }
func (l *slogLogger) Info(msg string, args ...any)    { l.logger.Info(msg, args...) }
func (l *slogLogger) Success(msg string, args ...any) { l.logger.Info(msg, args...) } // Success is logged as Info
func (l *slogLogger) Warn(msg string, args ...any)    { l.logger.Warn(msg, args...) }
func (l *slogLogger) Error(msg string, err error, args ...any) {
	allArgs := append(args, "err", err)
	l.logger.Error(msg, allArgs...)
}

// --- Simple Dumping Method ---
func (l *slogLogger) Dump(a ...any) {
	args := make([]any, 0, len(a)*2)
	for i, v := range a {
		key := fmt.Sprintf("dump_%d", i+1)
		args = append(args, key, v)
	}
	l.logger.Debug("Data dump", args...)
}

// --- Highlight Method ---
func (l *slogLogger) Highlight(color string, msg string, data ...any) {
	args := make([]any, 0, len(data)*2+2)
	args = append(args, "highlight_color", color)
	for i, v := range data {
		key := fmt.Sprintf("dump_%d", i+1)
		args = append(args, key, v)
	}
	l.logger.Debug(msg, args...)
}
