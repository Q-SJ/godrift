package log

import (
	"github.com/op/go-logging"
	"os"
	"strings"
)

var Logger *logging.Logger
var format logging.Formatter
var backend2Leveled, backend1Leveled logging.LeveledBackend

func init() {
	Logger = logging.MustGetLogger("main")
	format = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfile} %{shortfunc} ▶ %{level:.4s} %{color:reset} %{message}`,
	)
	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)

	// For messages written to backend2 we want to add some additional
	// information to the output, including the used log level and the name of
	// the function.
	backend2Formatter := logging.NewBackendFormatter(backend2, format)

	// Only errors and more severe messages should be sent to backend1
	backend1Leveled = logging.AddModuleLevel(backend1)
	backend2Leveled = logging.AddModuleLevel(backend2Formatter)

	backend1Leveled.SetLevel(logging.ERROR, "")

	// Set the backends to be used.
}

func SetLoggerLevel(level string) {
	if strings.EqualFold("DEBUG", level) {
		backend2Leveled.SetLevel(logging.DEBUG, "")
	} else if strings.EqualFold("INFO", level) {
		backend2Leveled.SetLevel(logging.INFO, "")
	} else {
		Logger.Warningf("The %s level is not supported, set the level to INFO.", level)
		backend2Leveled.SetLevel(logging.INFO, "")
	}
	logging.SetBackend(backend1Leveled, backend2Leveled)
}