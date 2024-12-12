package log

import (
	"errors"
	"fmt"
	"os"
)

type LevelName string

const (
	LevelNameDebug   LevelName = "DEBUG"
	LevelNameInfo    LevelName = "INFO"
	LevelNameWarning LevelName = "WARNING"
	LevelNameError   LevelName = "ERROR"
)

const (
	LOG_LEVEL_DEBUG   = iota
	LOG_LEVEL_INFO    = iota
	LOG_LEVEL_WARNING = iota
	LOG_LEVEL_ERROR   = iota
)

var verbosity = LOG_LEVEL_INFO

func SetVerbosity(newVerbosity int) {
	verbosity = newVerbosity
}

func SetVerbosityByName(levelName LevelName) {
	switch levelName {
	case LevelNameDebug:
		SetVerbosity(LOG_LEVEL_DEBUG)
	case LevelNameInfo:
		SetVerbosity(LOG_LEVEL_INFO)
	case LevelNameWarning:
		SetVerbosity(LOG_LEVEL_WARNING)
	case LevelNameError:
		SetVerbosity(LOG_LEVEL_ERROR)
	}
}

func Info(str string) {
	if verbosity >= LOG_LEVEL_INFO {
		fmt.Fprint(os.Stderr, str+"\n")
	}
}

func Infof(format string, args ...interface{}) {
	if verbosity <= LOG_LEVEL_INFO {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

func Debug(str string) {
	if verbosity <= LOG_LEVEL_DEBUG {
		fmt.Fprint(os.Stderr, str+"\n")
	}
}

func Debugf(format string, args ...interface{}) {

	if verbosity <= LOG_LEVEL_DEBUG {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

func Error(str string) {
	if verbosity <= LOG_LEVEL_ERROR {
		fmt.Fprint(os.Stderr, str+"\n")
	}
}

func WithError(err error) {
	fmt.Fprintf(os.Stderr, "%s", err)
}

// String is used both by fmt.Print and by Cobra in help text
func (e *LevelName) String() string {
	return string(*e)
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (e *LevelName) Set(v string) error {
	switch v {
	case string(LevelNameDebug), string(LevelNameInfo), string(LevelNameWarning), string(LevelNameError):
		*e = LevelName(v)
		return nil
	default:
		return errors.New(`must be one of "foo", "bar", or "moo"`)
	}
}

// Type is only used in help text
func (e *LevelName) Type() string {
	return "level"
}

func GetSupportedLevels() []string {
	return []string{string(LevelNameDebug), string(LevelNameInfo), string(LevelNameWarning), string(LevelNameError)}
}
