// pkg/logger/pretty_logger.go
package applogger

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"strings"
)

type prettyLogger struct{}

// NewPrettyLogger คือโรงงานสร้างนักข่าวสายสวยงาม
func NewPrettyLogger() Logger {
	return &prettyLogger{}
}

// print เป็นผู้ช่วยกลางที่ทำให้โค้ดไม่ซ้ำซ้อน
func (l *prettyLogger) print(color, emoji, level, msg string, args ...any) {
	location := getFileInfo()
	formattedArgs := formatArgs(args...)
	log.Printf("%s%s %-7s %s: %s%s%s", color, emoji, level, location, msg, formattedArgs, ColorReset)
}

// --- Structured Logging Methods ---
func (l *prettyLogger) Debug(msg string, args ...any) {
	l.print(ColorBlue, "🐛", "DEBUG", msg, args...)
}
func (l *prettyLogger) Info(msg string, args ...any) {
	l.print(ColorCyan, "ℹ️", "INFO", msg, args...)
}
func (l *prettyLogger) Success(msg string, args ...any) {
	l.print(ColorGreen, "✅", "SUCCESS", msg, args...)
}
func (l *prettyLogger) Warn(msg string, args ...any) {
	l.print(ColorYellow, "⚠️", "WARN", msg, args...)
}
func (l *prettyLogger) Error(msg string, err error, args ...any) {
	allArgs := append(args, "err", err)
	l.print(ColorRed, "❌", "ERROR", msg, allArgs...)
}

// --- Simple Dumping Method ---
func (l *prettyLogger) Dump(a ...any) {
	location := getFileInfo()
	var messages []string
	for _, v := range a {
		jsonBytes, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			messages = append(messages, fmt.Sprintf("<unmarshallable: %v>", err))
		} else {
			messages = append(messages, string(jsonBytes))
		}
	}
	log.Printf("%s🔍 DUMP  %s:\n%s%s", ColorPurple, location, strings.Join(messages, "\n"), ColorReset)
}

// --- Highlight Method ---
func (l *prettyLogger) Highlight(color string, msg string, data ...any) {
	location := getFileInfo()
	var messages []string
	for _, v := range data {
		jsonBytes, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			messages = append(messages, fmt.Sprintf("<unmarshallable: %v>", err))
		} else {
			messages = append(messages, string(jsonBytes))
		}
	}
	log.Printf("%s🎨 HIGHLIGHT %s: %s\n%s%s", color, location, msg, strings.Join(messages, "\n"), ColorReset)
}

// --- Helpers ---
func getFileInfo() string {
	for i := 3; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		if !strings.Contains(file, "pkg/logger") {
			parts := strings.Split(file, "/")
			return fmt.Sprintf("%s:%d", parts[len(parts)-1], line)
		}
	}
	return "???:0"
}

func formatArgs(args ...any) string {
	if len(args) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString(" |")
	for i := 0; i < len(args); i += 2 {
		key := args[i]
		var value any = "(MISSING)"
		if i+1 < len(args) {
			value = args[i+1]
		}
		builder.WriteString(fmt.Sprintf(" %s=%v", key, value))
	}
	return builder.String()
}
