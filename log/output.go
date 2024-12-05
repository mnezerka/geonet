package log

import (
	"fmt"
	"os"
)

var verbosity = 0

func SetVerbosity(newVerbosity int) {
	verbosity = newVerbosity
}

func Info(str string) {
	fmt.Fprint(os.Stderr, str+"\n")
}

func Infof(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func Debug(str string) {
	fmt.Fprint(os.Stderr, str+"\n")
}

func Debugf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func WithError(err error) {
	fmt.Fprintf(os.Stderr, "%s", err)
}
