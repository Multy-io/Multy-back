// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package slog

import (
	"strings"
	"sync"

	"github.com/jekabolt/slf"
)

const (
	// ContextField defines the field name to store context.
	ContextField = "context"
	rootLevelKey = "root"
)

// LogFactory extends the SLF LogFactory interface with a series of methods specific to the slog
// implementation.
type LogFactory interface {
	slf.LogFactory
	SetLevel(level slf.Level, contexts ...string)
	SetCallerInfo(callerInfo slf.CallerInfo, contexts ...string)
	AddEntryHandler(handler EntryHandler)
	SetEntryHandlers(handlers ...EntryHandler)
	Contexts() map[string]slf.StructuredLogger
	SetConcurrent(conc bool)
}

// New constructs a new logger conforming with SLF.
func New() LogFactory {
	res := &logFactory{
		root: rootLogger{
			minlevel: slf.LevelInfo,
			// essentially undefined: every logger that does not have it set (even to None)
			// should consult the root logger
			caller: slf.CallerInfo(-1),
		},
		contexts:   make(map[string]*logger),
		concurrent: false,
	}
	res.root.factory = res
	return res
}

// factory implements the slog.Logger interface.
type logFactory struct {
	sync.RWMutex
	root       rootLogger
	contexts   map[string]*logger
	handlers   []EntryHandler
	concurrent bool
}

// WithContext delivers a logger for the given context (reusing loggers for the same context).
func (lf *logFactory) WithContext(context string) slf.StructuredLogger {
	lf.RLock()
	ctx, ok := lf.contexts[context]
	lf.RUnlock()
	if ok {
		return ctx
	}
	fields := make(map[string]interface{})
	fields[ContextField] = context
	ctx = &logger{
		rootLogger: &rootLogger{minlevel: lf.root.minlevel, factory: lf.root.factory, caller: lf.root.caller},
		fields:     fields,
		caller:     lf.root.caller,
	}
	lf.Lock()
	lf.contexts[context] = ctx
	lf.Unlock()
	return ctx
}

// SetLevel sets the logging slf.Level to given contexts, all loggers if no context given, or the root
// logger when context defined as "root".
func (lf *logFactory) SetLevel(level slf.Level, contexts ...string) {
	// set on all current and root
	if len(contexts) == 0 {
		lf.root.minlevel = level
		lf.Lock()
		for _, logger := range lf.contexts {
			logger.rootLogger.minlevel = level
		}
		lf.Unlock()
		return
	}
	// setting on given only
	for _, context := range contexts {
		if strings.ToLower(context) != rootLevelKey {
			logger, _ := lf.WithContext(context).(*logger) // locks internally
			logger.rootLogger.minlevel = level
		} else {
			lf.root.minlevel = level
		}
	}
}

// SetCallerInfo sets the logging slf.CallerInfo to given contexts, all loggers if no context given,
// or the root logger when context defined as "root".
func (lf *logFactory) SetCallerInfo(callerInfo slf.CallerInfo, contexts ...string) {
	// set on all current and root
	if len(contexts) == 0 {
		lf.root.caller = callerInfo
		lf.Lock()
		for _, logger := range lf.contexts {
			logger.rootLogger.caller = callerInfo
		}
		lf.Unlock()
		return
	}
	// setting on given only
	for _, context := range contexts {
		if strings.ToLower(context) != rootLevelKey {
			logger, _ := lf.WithContext(context).(*logger) // locks internally
			logger.rootLogger.caller = callerInfo
		} else {
			lf.root.caller = callerInfo
		}
	}
}

// AddEntryHandler adds a handler for log entries that are logged at or above the set
// log slf.Level.
func (lf *logFactory) AddEntryHandler(handler EntryHandler) {
	lf.Lock()
	lf.handlers = append(lf.handlers, handler)
	lf.Unlock()
}

// SetEntryHandlers overwrites existing entry handlers with a new set.
func (lf *logFactory) SetEntryHandlers(handlers ...EntryHandler) {
	lf.Lock()
	lf.handlers = append([]EntryHandler{}, handlers...)
	lf.Unlock()
}

// Contexts returns all defined root logging contexts.
func (lf *logFactory) Contexts() map[string]slf.StructuredLogger {
	res := make(map[string]slf.StructuredLogger)
	lf.RLock()
	for key, val := range lf.contexts {
		res[key] = val
	}
	lf.RUnlock()
	return res
}

// SetConcurrent toggles concurrency in handling log messages. If concurrent, the output sequence
// of entries is not guaranteed to be the same as log entries input sequence, although the
// timestamp will correspond the time of logging, not handling. Default: not concurrent.
func (lf *logFactory) SetConcurrent(conc bool) {
	lf.concurrent = conc
}
