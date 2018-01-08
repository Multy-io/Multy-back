
[![Build status][buildimage]][build] [![Coverage][codecovimage]][codecov] [![GoReportCard][cardimage]][card] [![API documentation][docsimage]][docs]

# The reference [SLF][slf] implementation for Go 

The module `ventu-io/slog` provides a reference implementation of the structured logging facade ([SLF][slf]) for Go. The following are the main characteristics of this implementation:

* log levels can be set per context, to the root context or to all context;
* defines generic `Entry` and `EntryHandler` interfaces enabling adding arbitrary handlers;
* permits concurrent (default) or sequential processing of each log entry by each entry handler;
* defines a basic entry handler for logging into text files or terminal, which is fully parametrisable via a template (via the standard Go `text/template`)
* defines a JSON log entry handler for formatting JSON into a consumer (`io.Writer`)
* delivers about 1mil log entries to log entry handlers on conventional hardware concurrently or sequentially
* handles locking of contexts and handlers

More handlers will follow in due course.

## The factory API

The factory API is fairly straightforward and self explanatory. It is based on the `slf.LogFactory` adding just a few convenience method to set the level and define handers:

    type LogFactory interface {
        slf.LogFactory

        // SetLevel sets the logging slf.Level to given contexts, all loggers if no 
        // context given, or the root logger when context defined as "root".
        SetLevel(level slf.Level, contexts ...string)

        // SetCallerInfo sets the logging slf.CallerInfo to given contexts, all loggers if no context given,
        // or the root logger when context defined as "root".
        SetCallerInfo(callerInfo slf.CallerInfo, contexts ...string)

        // AddEntryHandler adds a handler for log entries that are logged at or above 
        // the set log slf.Level.
        AddEntryHandler(handler EntryHandler)

        // SetEntryHandlers sets a collection of handlers for log entries that are logged 
        // at or above the set log slf.Level.
        SetEntryHandlers(handlers ...EntryHandler)

        // Contexts retruns the currently defined collection of context loggers.
        Contexts() map[string]slf.StructuredLogger

        // SetConcurrent toggles concurrent execution of handler methods on log entries. 
        // Default is to log each entry non-concurrently, one after another.
        SetConcurrent(conc bool)
    }

## Usage 

This covers the initialisation only, otherwise see [slf]. At the application initialisation:

    func init() {
        // define a basic stderr log entry handler
        bh := basic.New()
        // optionally define the format (this here is the default one)
        bh.SetTemplate("{{.Time}} [\033[{{.Color}}m{{.Level}}\033[0m] {{.Context}}{{if .Caller}} ({{.Caller}}){{end}}: {{.Message}}{{if .Error}} (\033[31merror: {{.Error}}\033[0m){{end}} {{.Fields}}")


				// initialise and configure the SLF implementation
        lf := slog.New()
        // set common log level to INFO
        lf.SetLevel(slf.LevelInfo)
        // set log level for specific contexts to DEBUG
        lf.SetLevel(slf.LevelDebug, "app.package1", "app.package2")
        lf.AddEntryHandler(bh)
        lf.AddEntryHandler(json.New(os.Stderr))

        // make this into the one used by all the libraries
        slf.Set(lf) 
    }

## Output of the basic and json handlers


Given the above setup the log output of the application can look like this:

* for the JSON logger:

        {
          "timestamp": "2016-03-26T17:41:14.5517",
          "level": "WARN",
          "message": "Error while subscribing. Retrying in 30s",
          "error": "read: connection reset by peer",
          "fields": {
            "context": "probe.agent.task.Subscribe"
          }
        } 

* for the basic text logger (with coloured INFO if output to a terminal):

        17:41:14.551 [WARN] probe.agent.task.Subscribe: Error while subscribing. Retrying in 30s (error: read: connection reset by peer)

	 ![Basic output example][coloured]

##  Changelog

* 26.03.2016: Initial release
* 30.04.2016:
    * API: Added `Fatal` and `Fatalf` (matches SLF)
    * API: Added `SetCallerInfo` as the top-level initialization for all or selected loggers
    * Behaviour: `concurrent=false` by default
    * Fix: `Log(LevelPanic, ...)` triggers panic just as `Panic` and `Panicf`
    * Fix: Stopped JSON handler from outputting `error: "null"` on no error


## License

Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors.

Distributed under a MIT style license found in the [LICENSE][license] file.


[docs]: https://godoc.org/github.com/KristinaEtc/slog
[docsimage]: http://img.shields.io/badge/godoc-reference-blue.svg?style=flat

[build]: https://travis-ci.org/KristinaEtc/slog
[buildimage]: https://travis-ci.org/KristinaEtc/slog.svg?branch=master

[codecov]: https://codecov.io/github/KristinaEtc/slog?branch=master
[codecovimage]: https://codecov.io/github/KristinaEtc/slog/coverage.svg?branch=master

[card]: http://goreportcard.com/report/KristinaEtc/slog
[cardimage]: https://goreportcard.com/badge/github.com/KristinaEtc/slog

[license]: https://github.com/KristinaEtc/slog/blob/master/LICENSE

[slf]: https://github.com/KristinaEtc/slf
[coloured]: https://raw.githubusercontent.com/KristinaEtc/slog/master/basic/coloured-basic-output.png
