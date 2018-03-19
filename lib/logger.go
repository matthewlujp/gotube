package gotube

import (
	"log"
	"os"
)

type errorLogger struct {
	deploy bool
	logger *log.Logger
}

func newLogger(deploy bool) *errorLogger {
	return &errorLogger{
		deploy: deploy,
		logger: log.New(os.Stdout, "gotube >", 0),
	}
}

func (l *errorLogger) print(message string) {
	l.logger.Print(message)
}

func (l *errorLogger) printf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}

func (l *errorLogger) fatal(message string) {
	l.logger.Fatal(message)
}
