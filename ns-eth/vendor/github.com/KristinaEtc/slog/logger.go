// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package slog

import (
	"errors"
	"fmt"
	"github.com/jekabolt/slf"
	stdlog "log"
	"os"
	"path"
	"runtime"
	"time"
)

const (
	// TraceField defined the key for the field to store trace duration.
	TraceField = "trace"

	// CallerField defines the key for the caller information.
	CallerField = "caller"

	// ErrorField can be used by handlers to represent the error in the data field collection.
	ErrorField = "error"

	traceMessage = "trace"
)

var (
	noop  = &slf.Noop{}
	epoch = time.Time{}

	// ExitProcessor is executed on Log(LevelFatal) to terminate the application.
	ExitProcessor = func(message string) {
		os.Exit(1)
	}
)

// rootLogger represents a root logger for a context, all other loggers in the same context
// (with different fields) contain this one to identify the log level and entry handlers.
type rootLogger struct {
	minlevel slf.Level
	factory  *logFactory
	caller   slf.CallerInfo
}

// logger represents a logger in the context. It is created from the rootlogger by copying its
// fields. "Fields" access is not directly synchronised because fields are written on copy only,
// however it is synchronised indirectly to guarantee timestamp for tracing.
type logger struct {
	*rootLogger
	// not synced because ro outside of construction in with*
	fields map[string]interface{}
	caller slf.CallerInfo
	err    error
	// not synced
	lasttouch time.Time
	lastlevel slf.Level
}

// WithField implements the Logger interface.
func (log *logger) WithField(key string, value interface{}) slf.StructuredLogger {
	res := log.copy()
	res.fields[key] = value
	return res
}

// WithFields implements the Logger interface.
func (log *logger) WithFields(fields slf.Fields) slf.StructuredLogger {
	res := log.copy()
	for k, v := range fields {
		res.fields[k] = v
	}
	return res
}

// WithCaller implements the Logger interface.
func (log *logger) WithCaller(caller slf.CallerInfo) slf.StructuredLogger {
	res := log.copy()
	res.caller = caller
	return res
}

// WithError implements the Logger interface.
func (log *logger) WithError(err error) slf.Logger {
	res := log.copy()
	res.err = err
	return res
}

// Log implements the Logger interface.
func (log *logger) Log(level slf.Level, message string) slf.Tracer {
	return log.log(level, message)
}

// Trace implements the Logger interface.
func (log *logger) Trace(err *error) {
	lasttouch := log.lasttouch
	level := log.lastlevel
	if lasttouch != epoch && level >= log.rootLogger.minlevel {
		var entry *entry
		if err != nil {
			entry = log.entry(level, traceMessage, 2, *err)
		} else {
			entry = log.entry(level, traceMessage, 2, nil)
		}
		entry.fields[TraceField] = time.Now().Sub(lasttouch)
		log.handleall(entry)
	}
	log.lasttouch = epoch
}

// Debug implements the Logger interface.
func (log *logger) Debug(message string) slf.Tracer {
	return log.log(slf.LevelDebug, message)
}

// Debugf implements the Logger interface.
func (log *logger) Debugf(format string, args ...interface{}) slf.Tracer {
	return log.logf(format, slf.LevelDebug, args...)
}

// Info implements the Logger interface.
func (log *logger) Info(message string) slf.Tracer {
	return log.log(slf.LevelInfo, message)
}

// Infof implements the Logger interface.
func (log *logger) Infof(format string, args ...interface{}) slf.Tracer {
	return log.logf(format, slf.LevelInfo, args...)
}

// Warn implements the Logger interface.
func (log *logger) Warn(message string) slf.Tracer {
	return log.log(slf.LevelWarn, message)
}

// Warnf implements the Logger interface.
func (log *logger) Warnf(format string, args ...interface{}) slf.Tracer {
	return log.logf(format, slf.LevelWarn, args...)
}

// Error implements the Logger interface.
func (log *logger) Error(message string) slf.Tracer {
	return log.log(slf.LevelError, message)
}

// Errorf implements the Logger interface.
func (log *logger) Errorf(format string, args ...interface{}) slf.Tracer {
	return log.logf(format, slf.LevelError, args...)
}

// Panic implements the Logger interface.
func (log *logger) Panic(message string) {
	log.log(slf.LevelPanic, message)
}

// Panicf implements the Logger interface.
func (log *logger) Panicf(format string, args ...interface{}) {
	log.logf(format, slf.LevelPanic, args...)
}

// Fatal implements the Logger interface.
func (log *logger) Fatal(message string) {
	log.log(slf.LevelFatal, message)
}

// Fatalf implements the Logger interface.
func (log *logger) Fatalf(format string, args ...interface{}) {
	log.logf(format, slf.LevelFatal, args...)
}

// Log implements the Logger interface.
func (log *logger) log(level slf.Level, message string) slf.Tracer {
	if level < log.rootLogger.minlevel {
		return noop
	}
	return log.checkedlog(level, message)
}

func (log *logger) logf(format string, level slf.Level, args ...interface{}) slf.Tracer {
	if level < log.rootLogger.minlevel {
		return noop
	}
	message := fmt.Sprintf(format, args...)
	return log.checkedlog(level, message)
}

func (log *logger) checkedlog(level slf.Level, message string) slf.Tracer {
	log.handleall(log.entry(level, message, 4, log.err))
	log.lasttouch = time.Now()
	log.lastlevel = level
	if level == slf.LevelPanic {
		panic(errors.New(message))
	} else if level == slf.LevelFatal {
		ExitProcessor(message)
	}
	return log
}

func (log *logger) copy() *logger {
	res := &logger{
		rootLogger: log.rootLogger,
		fields:     make(map[string]interface{}),
		caller:     log.caller,
	}
	for key, value := range log.fields {
		res.fields[key] = value
	}
	return res
}

func (log *logger) entry(level slf.Level, message string, skip int, err error) *entry {
	fields := make(map[string]interface{})
	for key, value := range log.fields {
		fields[key] = value
	}
	caller := log.caller
	if caller < slf.CallerNone {
		caller = log.rootLogger.caller
	}
	if caller == slf.CallerLong || caller == slf.CallerShort {
		if _, file, line, ok := runtime.Caller(skip); ok {
			if caller == slf.CallerShort {
				file = path.Base(file)
			}
			fields[CallerField] = fmt.Sprintf("%s:%d", file, line)
		}
	}
	return &entry{tm: time.Now(), level: level, message: message, err: err, fields: fields}
}

func (log *logger) handleall(entry *entry) {
	f := log.rootLogger.factory
	f.RLock()
	handlers := make([]EntryHandler, len(f.handlers))
	copy(handlers, f.handlers)
	f.RUnlock()

	for _, handler := range handlers {
		if log.factory.concurrent {
			go log.handleone(handler, entry)
		} else {
			log.handleone(handler, entry)
		}
	}
	if log.factory.concurrent {
		runtime.Gosched()
	}
}

func (log *logger) handleone(h EntryHandler, e *entry) {
	if err := h.Handle(e); err != nil {
		// fall back to standard logging to output entry handler error
		stdlog.Printf("log handler error: %v\n", err.Error())
	}
}
