//+build debug

package debug

import (
	"log"
	"os"
	"strconv"

	"github.com/davecgh/go-spew/spew"
)

// Enabled is a global flag to fold blocks of code
const Enabled = true

// LogEnabled is variable that users can initialize via environment
// variable CURSOR_DEBUG_LOG. If this is false, logs will not be
// generated even in the presense of `-tags debug`. By default
// this is false
var LogEnabled = false
var logger *log.Logger

func init() {
	LogEnabled, _ = strconv.ParseBool(os.Getenv("CURSOR_DEBUG_LOG"))
	logger = log.New(os.Stdout, "|DEBUG| ", 0)
}

// Printf prints debug messages. Only available if compiled with "debug" tag
func Printf(f string, args ...interface{}) {
	if !LogEnabled {
		return
	}
	logger.Printf(f, args...)
}

func Dump(v ...interface{}) {
	if !LogEnabled {
		return
	}
	spew.Dump(v...)
}
