package utils
import (
	"fmt"
	"os"
)

const (
	EXITCODE_OK = 0
	EXITCODE_MISCONFIGURATION = 1
	EXITCODE_RUNTIME_ERROR = 2
	EXITCODE_SHUTDOWN_FORCED = 3
)

func Quit(code int) {
	os.Exit(code)
}

func Quitf(code int, msg string, args... interface{}) {
	fmt.Println(fmt.Sprintf("Unknown queue type requested: %s", args...))
	Quit(code)
}
