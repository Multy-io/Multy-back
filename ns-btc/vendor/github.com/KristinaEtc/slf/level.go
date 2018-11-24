// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package slf

import (
	"fmt"
	"strings"
)

// Level represents log level of the structured logger.
type Level int

// Log level constants.
const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelPanic
	LevelFatal
)

func (l Level) String() string {
	s, err := l.string()
	if err == nil {
		return s
	}
	return "<NA>"
}

func (l Level) string() (string, error) {
	switch l {
	case LevelDebug:
		return "DEBUG", nil
	case LevelInfo:
		return "INFO", nil
	case LevelWarn:
		return "WARN", nil
	case LevelError:
		return "ERROR", nil
	case LevelPanic:
		return "PANIC", nil
	case LevelFatal:
		return "FATAL", nil
	default:
		return "", fmt.Errorf("slf: unknown level %d", int(l))
	}
}

// MarshalJSON provides a JSON representation of the log level.
func (l Level) MarshalJSON() ([]byte, error) {
	s, err := l.string()
	if err == nil {
		return []byte(`"` + s + `"`), nil
	}
	return nil, err
}

// UnmarshalJSON parses the JSON representation of the log level into a Level object.
func (l *Level) UnmarshalJSON(data []byte) error {
	s := strings.ToLower(string(data))
	switch s {
	case `"debug"`:
		fallthrough
	case "debug":
		*l = LevelDebug
	case `"info"`:
		fallthrough
	case "info":
		*l = LevelInfo
	case `"warn"`:
		fallthrough
	case "warn":
		*l = LevelWarn
	case `"error"`:
		fallthrough
	case "error":
		*l = LevelError
	case `"panic"`:
		fallthrough
	case "panic":
		*l = LevelPanic
	case `"fatal"`:
		fallthrough
	case "fatal":
		*l = LevelFatal
	default:
		return fmt.Errorf("slf: unknown level %v", s)
	}
	return nil
}
