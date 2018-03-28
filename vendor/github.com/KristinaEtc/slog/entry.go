// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package slog

import (
	"github.com/KristinaEtc/slf"
	"time"
)

// EntryHandler processes filtered log entries in independent go-routines.
type EntryHandler interface {

	// Handle processes a filtered log entry (must not write to the entry field map, which is
	// read concurrently by all handlers).
	Handle(Entry) error
}

// Entry represents a log entry for structured logging. Entries are only created when the requested
// level is same or above the minimum log level of the context root.
type Entry interface {

	// Time represents entry time stamp.
	Time() time.Time

	// Level represents the log level.
	Level() slf.Level

	// Message reresents the log formatted message.
	Message() string

	// Error, if present, represents the error to be logged along with the message and the fields.
	Error() error

	// Fields represents structured log information.
	Fields() map[string]interface{}
}

type entry struct {
	tm      time.Time
	level   slf.Level
	message string
	err     error
	fields  map[string]interface{}
}

func (e *entry) Time() time.Time {
	return e.tm
}

func (e *entry) Level() slf.Level {
	return e.level
}

func (e *entry) Message() string {
	return e.message
}

func (e *entry) Error() error {
	return e.err
}

func (e *entry) Fields() map[string]interface{} {
	return e.fields
}
