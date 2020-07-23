package log

import (
	"fmt"

	"github.com/go-logr/logr"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// Log logs messages to various levels.
//
// Based on the go-logr package.
//
// The Log type enables various verbosity levels to be logged
// based on to the --zap-level argument to the operator.
//
// To create an instance of the log named "main":
//
//	log := log.NewLog("main")
//
// When --zap-level is not provided
// then the logging defaults to INFO level where log.Warning() and log.Info() are logged,
//
// if --zap-level is set to 1
// then the logging goes to the DEBUG level where log.Debug() is logged,
//
// while the --zap-level set to 2 is reserved for a finer DEBUG-level logging
// provided by the log.Trace() function.
type Log struct {
	log *logr.Logger //log instance
}

// Error logs the message using go-logr package on a default level as ERROR
func (l *Log) Error(err error, msg string, keysAndValues ...interface{}) {
	(*l.log).Error(err, msg, keysAndValues...)
}

// Warning logs the message using go-logr package on a default level as WARNING
func (l *Log) Warning(msg string, keysAndValues ...interface{}) {
	if (*l.log).V(0).Enabled() {
		(*l.log).V(0).Info(fmt.Sprintf("WARNING: %s", msg), keysAndValues...)
	}
}

// Info logs the message using go-logr package on a default level as INFO
func (l *Log) Info(msg string, keysAndValues ...interface{}) {
	if (*l.log).V(0).Enabled() {
		(*l.log).V(0).Info(msg, keysAndValues...)
	}
}

// Debug logs the message using go-logr package on a V=1 level as DEBUG
func (l *Log) Debug(msg string, keysAndValues ...interface{}) {
	if (*l.log).V(1).Enabled() {
		(*l.log).V(1).Info(msg, keysAndValues...)
	}
}

// Trace logs the message using go-logr package on a V=1 level as TRACE
func (l *Log) Trace(msg string, keysAndValues ...interface{}) {
	if (*l.log).V(2).Enabled() {
		// The V(1) level here is intentional:
		//   sTRACE = finer DEBUG logging but is only enabled by setting V=2
		(*l.log).V(1).Info(fmt.Sprintf("TRACE: %s", msg), keysAndValues...)
	}
}

// NewLog returns an instance of a Log
func NewLog(name string, keysAndValues ...interface{}) *Log {
	log := logf.Log.WithName(name).WithValues(keysAndValues...)
	l := &Log{
		log: &log,
	}
	return l
}

// SetLog sets a concrete logging implementation for all deferred Loggers.
func SetLog(log logr.Logger) {
	logf.SetLogger(log)
}

// WithValues adds some key-value pairs of context to a log.
// See Info for documentation on how key/value pairs work.
func (l *Log) WithValues(keysAndValues ...interface{}) *Log {
	lgr := ((*l.log).WithValues(keysAndValues...))
	log := &Log{
		log: &lgr,
	}
	return log
}

// WithName adds a new element to the log's name. Successive calls with WithName continue to append
// suffixes to the log's name. It's strongly reccomended that name segments contain only letters,
// digits, and hyphens (see the package documentation for more information).
func (l *Log) WithName(name string) *Log {
	lgr := ((*l.log).WithName(name))
	log := &Log{
		log: &lgr,
	}
	return log
}
