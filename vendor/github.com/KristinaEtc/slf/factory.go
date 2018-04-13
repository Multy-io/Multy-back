// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package slf

var factory LogFactory = &Noop{}

// IsSet checks if the global log factory is set and not Noop.
func IsSet() bool {
	if _, ok := factory.(*Noop); !ok {
		return true
	}
	return false
}

// Set sets the global log factory.
func Set(log LogFactory) {
	if log != nil {
		factory = log
	}
}

// WithContext returns a logger with context set to the given string.
func WithContext(context string) StructuredLogger {
	return factory.WithContext(context)
}
