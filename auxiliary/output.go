package auxiliary

import (
	"fmt"
	"os"
)

func Info(str string) {
	fmt.Fprint(os.Stdout, str+"\n")
}

func Infof(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format+"\n", args...)
}

func Debug(str string) {
	fmt.Fprint(os.Stderr, str+"\n")
}

func Debugf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
