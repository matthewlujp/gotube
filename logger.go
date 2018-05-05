package gotube

import (
	"log"
	"os"
)

type errorLogger struct {
	debug  bool
	logger *log.Logger
}

func newLogger(deploy bool) *errorLogger {
	return &errorLogger{
		debug:  deploy,
		logger: log.New(os.Stderr, "gotube >", 0),
	}
}

func (l *errorLogger) print(message string) {
	if l.debug {
		l.logger.Print(message)
	}
}

func (l *errorLogger) printf(format string, v ...interface{}) {
	if l.debug {
		l.logger.Printf(format, v...)
	}
}

func (l *errorLogger) fatal(message string) {
	if l.debug {
		l.logger.Fatal(message)
	}
}
