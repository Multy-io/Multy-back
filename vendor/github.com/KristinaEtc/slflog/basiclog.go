// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

// Package basic provides a basic text/terminal log entry handler for slog. It allows to define
// the format of the output string (via Go templates), the format of date and time, the colouring
// of entries in case of terminal output as well as the writer to redirect the output to. The
// handler synchronises on write to allow the same handler to be used from concurrent routines.
package slflog

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"text/template"

	"github.com/KristinaEtc/slf"
	"github.com/KristinaEtc/slog"
)

const (
	red    = 31
	green  = 32
	yellow = 33
	blue   = 34
	gray   = 37
)

const (
	// StandardTermTemplate represents a standard template for terminal output (default).
	StandardTermTemplate = "{{.Time}} [\033[{{.Color}}m{{.Level}}\033[0m] {{.Context}}{{if .Caller}} ({{.Caller}}){{end}}: {{.Message}}{{if .Error}} (\033[31merror: {{.Error}}\033[0m){{end}} {{.Fields}}"

	// StandardTextTemplate represents a standard template for text file output or any other writers
	// not supporting terminal colouring.
	StandardTextTemplate = "{{.Time}} [{{.Level}}] {{.Context}}{{if .Caller}} ({{.Caller}}){{end}}: {{.Message}}{{if .Error}} (error: {{.Error}}){{end}} {{.Fields}}"

	// StandardTimeFormat represents the time format used in the handler by default.
	// ISO 8601
	StandardTimeFormat = "2006-01-02/15:04:05.000"
)

// Handler represents a log entry handler capable of formatting structured log data into
// a text format (text files, stderr with or without colouring etc).
type Handler struct {
	sync.Mutex
	writer        io.Writer
	colors        map[slf.Level]int
	timeFormatStr string
	templateStr   string
	template      *template.Template
	Level         slf.Level
}

// New constructs a new handler with default template, time formatting, colours and stderr as
// output.
func New(level ...slf.Level) *Handler {

	var logTemplate string
	if runtime.GOOS == "windows" {
		logTemplate = StandardTextTemplate
	} else {
		logTemplate = StandardTermTemplate
	}

	res := &Handler{
		writer:        os.Stderr,
		colors:        make(map[slf.Level]int),
		templateStr:   logTemplate,
		timeFormatStr: StandardTimeFormat,
		template:      template.Must(template.New("entry").Parse(logTemplate + "\n")),
	}
	if level == nil {
		// TODO: CHANGE TO ROOTLEVEL
		res.Level = slf.LevelInfo
	} else {
		// TODO: CHECK LOGLEVEL FOR VALID
		res.Level = level[0]
	}
	res.colors[slf.LevelDebug] = blue
	res.colors[slf.LevelInfo] = green
	res.colors[slf.LevelWarn] = yellow
	res.colors[slf.LevelError] = red
	res.colors[slf.LevelPanic] = red
	return res
}

func ConfigFileOutput(logHandlers *[]slog.EntryHandler, lvl slf.Level, fileName string) error {

	bHandler := New(lvl)
	logfd, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[basiclog.go] Error: could not open/create logfile: %s %v\n", fileName, err)
		return err
	}

	bHandler.SetWriter(logfd)
	*logHandlers = append(*logHandlers, bHandler)
	return nil
}

func ConfigWriterOutput(logHandlers *[]slog.EntryHandler, lvl slf.Level, fd io.Writer) {
	bHandler := New(lvl)
	bHandler.SetWriter(fd)
	*logHandlers = append(*logHandlers, bHandler)
}

// SetWriter defines the writer to use to output log strings (default: stderr).
func (h *Handler) SetWriter(w io.Writer) {
	h.writer = w
}

// SetTemplate defines the formatting of the log string using the standard Go template syntax.
// See the Data structure for the definition of all supported template fields.
func (h *Handler) SetTemplate(s string) error {
	t, err := template.New("entry").Parse(s + "\n")
	if err != nil {
		return err
	}
	h.templateStr = s
	h.template = t
	return nil
}

// SetTimeFormat defines the formatting of time used for output into the template.
func (h *Handler) SetTimeFormat(f string) {
	h.timeFormatStr = f
}

// SetColors overwrites the level-colour mapping. Every missing mapping will be replaced by gray.
func (h *Handler) SetColors(colors map[slf.Level]int) {
	// no validation: if color not found, gray is used
	h.colors = colors
}

// Handle outputs a textual representation of the log entry into a text writer (stderr, file etc.).
func (h *Handler) Handle(e slog.Entry) (err error) {

	if e.Level() < h.Level {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	d := &Data{
		Time:    e.Time().Format(h.timeFormatStr),
		Level:   e.Level().String(),
		Context: h.contextstring(e),
		Message: e.Message(),
		Error:   e.Error(),
		Caller:  h.callerstring(e),
		Fields:  h.fieldstring(e),
		Color:   h.color(e),
	}
	h.Lock()
	defer h.Unlock()
	err = h.template.Execute(h.writer, d)
	return err
}

// Data supplies log data to the template formatter for outputting into the log string. This
// structure defines all the fields that can be used in the template.
type Data struct {
	Time    string
	Level   string
	Context string
	Message string
	Error   error
	Caller  string
	Fields  string
	Color   int
}

func Close() {
	//...
}

func (h *Handler) contextstring(e slog.Entry) string {
	return fmt.Sprint(e.Fields()[slog.ContextField])
}

func (h *Handler) callerstring(e slog.Entry) string {
	c, ok := e.Fields()[slog.CallerField]
	if ok {
		return fmt.Sprint(c)
	}
	return ""
}

func (h *Handler) color(e slog.Entry) int {
	c, ok := h.colors[e.Level()]
	if ok {
		return c
	}
	return gray
}

func (h *Handler) fieldstring(e slog.Entry) string {
	fs := []field{}
	for key, value := range e.Fields() {
		if key == slog.ContextField && strings.Contains(h.templateStr, "{{.Context}}") {
			continue
		}
		if key == slog.CallerField && strings.Contains(h.templateStr, "{{.Caller}}") {
			continue
		}
		fs = append(fs, field{key, value})
	}
	if e.Error() != nil && !strings.Contains(h.templateStr, "{{.Error}}") {
		fs = append(fs, field{slog.ErrorField, e.Error().Error()})
	}

	sort.Sort(sortablefields(fs))

	res := []string{}
	for _, f := range fs {
		res = append(res, fmt.Sprintf("%s=%v", f.key, f.value))
	}
	return strings.Join(res, "; ")
}

type field struct {
	key   string
	value interface{}
}

type sortablefields []field

func (sf sortablefields) Len() int {
	return len(sf)
}

func (sf sortablefields) Swap(i, j int) {
	sf[i], sf[j] = sf[j], sf[i]
}

func (sf sortablefields) Less(i, j int) bool {
	return sf[i].key < sf[j].key
}
