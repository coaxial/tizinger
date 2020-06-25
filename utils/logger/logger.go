package logger

import (
	"log"
	"os"
)

var (
	// Trace logs verbose debug/troubleshooting information.
	Trace *log.Logger
	// Info logs information that helps understand what the code is doing.
	Info *log.Logger
	// Warning logs information worth noting, but that doesn't prevent code
	// from running further.
	Warning *log.Logger
	// Error logs information about errors which prevent code from running
	// further.
	Error *log.Logger
)

func init() {
	flags := log.Ldate | log.LUTC | log.Lshortfile | log.Lmicroseconds | log.Lmsgprefix

	Trace = log.New(os.Stdout, "TRACE: ", flags)
	Info = log.New(os.Stdout, "INFO: ", flags)
	Warning = log.New(os.Stdout, "WARNING: ", flags)
	Error = log.New(os.Stdout, "ERROR: ", flags)
}
