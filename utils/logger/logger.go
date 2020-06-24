package logger

import (
	"log"
	"os"
)

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func init() {
	flags := log.Ldate | log.LUTC | log.Lshortfile | log.Lmicroseconds | log.Lmsgprefix

	Trace = log.New(os.Stdout, "TRACE: ", flags)
	Info = log.New(os.Stdout, "INFO: ", flags)
	Warning = log.New(os.Stdout, "WARNING: ", flags)
	Error = log.New(os.Stdout, "ERROR: ", flags)
}
