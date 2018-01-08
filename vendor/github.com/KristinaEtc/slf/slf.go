// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

// Package slf provides a Structured Log Facade for Go and factory functions to retrieve a
// a logger instance in code using this interface. Its use is analogous to slf4j in Java,
// with the difference that "s" stands here for structured rather than simple..
//
// The package not provide any actual logger implementation with the exception of the internal
// noop one (No Operation) delivered by default via the factory functions to permit using
// the interface without any further configuration. For a matching logger implementation see e.g.
// github.com/KristinaEtc/slog.
package slf

// CallerInfo defines an enumeration for the types of caller information to output, short or long.
type CallerInfo int

const (
	// CallerNone defines no caller info in the log.
	CallerNone CallerInfo = iota

	// CallerShort defines short caller info in the log (base name).
	CallerShort

	// CallerLong defines long caller info in the log (full path).
	CallerLong
)

// LogFactory represents a logger API for structured logging.
type LogFactory interface {
	// WithContext returns a logger with context set to a string.
	WithContext(string) StructuredLogger
}

// StructuredLogger represents a logger that can define a structured context by adding data fields..
type StructuredLogger interface {
	Logger

	// WithField adds a named data field to the logger context.
	WithField(string, interface{}) StructuredLogger

	// WithFields adds a number of named fields to the logger context.
	WithFields(Fields) StructuredLogger

	// WithCaller adds caller information to the data fields.
	WithCaller(CallerInfo) StructuredLogger

	// WithError adds an error record to the logger context (only one permitted).
	WithError(error) Logger
}

// Logger represents a generic leveled log interface.
type Logger interface {
	// Log logs the string with the given level.
	Log(Level, string) Tracer

	// Debug logs the string with the corresponding level.
	Debug(string) Tracer

	// Debugf formats and logs the string with the corresponding level.
	Debugf(string, ...interface{}) Tracer

	// Info logs the string with the corresponding level.
	Info(string) Tracer

	// Infof formats and logs the string with the corresponding level.
	Infof(string, ...interface{}) Tracer

	// Warn logs the string with the corresponding level.
	Warn(string) Tracer

	// Warnf formats and logs the string with the corresponding level.
	Warnf(string, ...interface{}) Tracer

	// Error logs the string with the corresponding level.
	Error(string) Tracer

	// Errorf formats and logs the string with the corresponding level.
	Errorf(string, ...interface{}) Tracer

	// Panic logs the string with the corresponding level and panics.
	Panic(string)

	// Panicf formats and logs the string with the corresponding level and panics.
	Panicf(string, ...interface{})

	// Fatal logs the string with the corresponding level and then calls os.Exit(1).
	Fatal(string)

	// Fatalf formats and logs the string with the corresponding level and then calls os.Exit(1).
	Fatalf(string, ...interface{})
}

// Tracer represents a logger that will trace the execution time since the last log event
// (common use case to call it in defer).
type Tracer interface {
	Trace(*error)
}

// Fields defines a field map.
type Fields map[string]interface{}
