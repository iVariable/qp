package utils

import (
	"fmt"
	"os"
)

// Exit codes
const (
	ExitCodeOk = 0
	ExitCodeMisconfiguration = 1
	ExitCodeRuntimeError = 2
	ExitCodeShutdownForced = 3
)

// Quit immediately stop program and quit with exit code
func Quit(code int) {
	os.Exit(code)
}

// Quitf immediately stop program and quit with exit code and message
func Quitf(code int, msg string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(msg, args...))
	Quit(code)
}
