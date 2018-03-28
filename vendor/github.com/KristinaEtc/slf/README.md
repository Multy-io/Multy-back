
[![Build status][buildimage]][build] [![Coverage][codecovimage]][codecov] [![GoReportCard][cardimage]][card] [![API documentation][docsimage]][docs]

# Structured Logging Facade (SLF) for Go

####   See also the reference implementation of this interface in [ventu-io/slog][slog].

Logging is central to any production quality code and there are many logging modules for Go. Surprisingly, none of the most commonly used ones provides a clean separation between the interface and the implementation of logging. If there are ones on the market, the author apparently has not found one.  Guaranteeing consistent logging at the library level for libraries coming from different development teams is, therefore, next to impossible.

Java developers have had the solution for a long while now: a combination of SLF4J (simple logging facade for java) and one of the implementations such as logback or log4j etc.

The `slf` module provides for Go what SLF4J does for Java: a clean logging facade with no implementation and a factory that any code can use to pull a logger configured at the start of the application. It then goes a step further and defines a structured logging facade rather than a "simple" one. 

The design adopts ideas from [apex/log][apexlog], which itself was inspired by [logrus]: both good modules for structured logging, yet failing exactly at the point of providing any generic interface to logging to which different implementations could be adopted. Generally speaking, this is not a substitute for either of the latter, they both can be adopted to work behind this interface. However, an independent implementation matching the interface is also provided separately in [ventu-io/slog][slog]. It provides context-specific log levels, concurrent and sequential log processing, a text/terminal log entry handler for coloured output with custom formatting of log entries using Go `text/template` and custom time format. At the moment it also provides a JSON log entry handler and more handlers can be added extrnally or will be added in due course.

##  The API

The API is fairly confined and straightforward, so it is worth looking into the code while adopting it. However, the general concept can be outlined as follows:

* an implementation is set once at the init of the main application by calling `slf.Set(logFactory)`;
* context-specific logger instances, e.g. those for a package, a struct or a module, are retrieved from the factory using `slf.WithContext(contextDef)`. The interface prescribes at least one structured field, `context`;
* these can then be used with fields to define the structure of log entries or even without any further fields, just like a plain vanilla logger. Arbitrary fields are accepted via `WithField("key", value)` or `WithFields(slf.Fields{"key": value})` and the predefined ones via `WithError(err)` or `WithCaller(slf.CallerShort)`;
* leveled log entries can be generated with or without a formatter, `logger.Debugf("%v", "message")` and `logger.Info("message")`;
* the interface defines 5 serializable log levels `DEBUG`, `INFO`, `WARN`, `ERROR`, `PANIC`, `FATAL` (`TRACE` is explicitly excluded);
* finally, each log entry can be followed by a `Trace` method, which is supposed to log execution time counting from the last log entry by the same logger using the same log level as the last log entry. Using it with `defer` provides a clean mechanism to trace the execution times of methods in one line.

The interface defines no structure of the actual log entry, nor the mechanism of how the log levels are set to the contexts (if at all), nor log entry handler interface, nor the actual log entry handlers. All these entities are what libraries using the logger do not care about and are only of concern for the application putting the libraries together. Therefore, all of these are deferred to the actual implementation used in the application (and are for example defined in the reference [slog] implementation).
 
## Examples

### Using within a library

As mentioned above, there is no need to bind any implementation of SLF into a library. The `slf` module provides a `Noop` implementation by default, that is the one with No-Operation, which will make sure your library does not panic on missing logging implementation. The factory methods and all the logging routines will work as is, yet no output will be generated. Two exceptions are the `Panic` command, which will panic even though no log output will be generated, and `Fatal`, which will call `os.Exit(1)` even though no log output will be generated.

One should be careful not to initialise any context logger, that is the one obtained with `slf.WithContext`, in the `init` or into a package-level variable directly, as this will return an instance of the `Noop` implementation before the factory could be reset to use an actual logging implementation in the main application. However, the following three approaches are all valid:

* define a package-level function to retrieve a context logger, e.g. 

        func logger() slf.StructuredLogger {
            return slf.WithContext("app.package")
        }
       
        func HandleRequest(req Request) (err error) {
            defer logger().WithField("request", req).Info("processing request").Trace(&err)
            return handle(req)
        }

* define a logger within a struct (the approach adopted by the author):

        type RequestHandler struct {
            logger slf.StructuredLogger
        } 
       
        func New() *RequestHandler {
            return &RequestHandler{
                logger: slf.WithContext("app.package"),
            }
        }

        func (rh *RequestHandler) Handle(req Request) (err error) {
            defer rh.logger.WithField("request", req).Info("processing request").Trace(&err)
            return handle(req)
        }

* finally, get a context logger directly at the point where you need it. The reference implementation reuses an earlier created logger for the context, however, it comes at a cost of verbosity of your code and going through locking of resources at the implementation level:

        func HandleRequest(req Request) (err error) {
            defer slf.WithContext("app.package").WithField("request", req).Info("processing request").Trace(err)
            return handle(req)
        }

Further examples for the use of the facade include:
 
* logging without any fields (except for `context`):
 
         logger.Debugf("done %v", "for now")
     
* deriving a new loggers in the same context with further added fields:
  
         logger1 := logger0.WithField("A", a).WithField("B", b)
         logger1.WithError(err).Warn("handling error")

* adding or supressing caller information without altering the parent logger

         logger1 := logger0.WithCaller(slf.CallerLong)
         logger2 := logger1.WithCaller(slf.CallerNone)
         logger1.Info("with caller")
         logger2.Info("without caller")

* using `Trace` accepts a pointer to an error. This is done so that the call can be defined in `defer`, which will get an optional error during the execution (nil is accepted).
 
### Using within an end-user application

The initialisation clause of an end-user application is where the actual implementation needs to be configured, and only here. The configuration could look like this if the reference [slog] implementation is used (for the sake of clarity). Other implementation will have their own initialisation logic, but the overall idea will remain the same:

    func init() {
        // define a basic stderr log entry handler, or any other log entry handler of choice
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



## License

Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors.

Distributed under a MIT style license found in the [LICENSE][license] file.


[docs]: https://godoc.org/github.com/KristinaEtc/slf
[docsimage]: http://img.shields.io/badge/godoc-reference-blue.svg?style=flat

[build]: https://travis-ci.org/KristinaEtc/slf
[buildimage]: https://travis-ci.org/KristinaEtc/slf.svg?branch=master

[codecov]: https://codecov.io/github/KristinaEtc/slf?branch=master
[codecovimage]: https://codecov.io/github/KristinaEtc/slf/coverage.svg?branch=master

[card]: http://goreportcard.com/report/KristinaEtc/slf
[cardimage]: https://goreportcard.com/badge/github.com/KristinaEtc/slf

[license]: https://github.com/KristinaEtc/slf/blob/master/LICENSE

[apexlog]: https://github.com/apex/log
[logrus]: https://github.com/Sirupsen/logrus
[slog]: https://github.com/KristinaEtc/slog


