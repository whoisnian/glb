// Package logger implements a simple logger for local debug.
package logger

import (
	"fmt"
	"log"
	"os"

	"github.com/whoisnian/glb/ansi"
)

var debug bool = false
var lout, lerr *log.Logger

var (
	tagI string = "[I]"
	tagW string = "[W]"
	tagE string = "[E]"
	tagD string = "[D]"
	tagR string = "[R]"
)

func init() {
	lout = log.New(os.Stdout, "", log.LstdFlags)
	lerr = log.New(os.Stderr, "", log.LstdFlags)
}

func SetDebug(value bool) {
	debug = value
	if debug {
		tagI = ansi.Green + tagI + ansi.Reset
		tagW = ansi.Yellow + tagW + ansi.Reset
		tagE = ansi.Red + tagE + ansi.Reset
		tagD = ansi.Magenta + tagD + ansi.Reset
		tagR = ansi.Blue + tagR + ansi.Reset
	}
}

func IsDebug() bool {
	return debug
}

func Info(v ...interface{}) {
	lout.Output(2, tagI+" "+fmt.Sprint(v...)+"\n")
}

func Warn(v ...interface{}) {
	lout.Output(2, tagW+" "+fmt.Sprint(v...)+"\n")
}

func Error(v ...interface{}) {
	lerr.Output(2, tagE+" "+fmt.Sprint(v...)+"\n")
}

// Example:
//   if logger.IsDebug() {
//       logger.Debug("The quick brown fox jumps over the lazy dog.")
//   }
func Debug(v ...interface{}) {
	lout.Output(2, tagD+" "+fmt.Sprint(v...)+"\n")
}

func Req(v ...interface{}) {
	lout.Output(2, tagR+" "+fmt.Sprint(v...)+"\n")
}

func Panic(v ...interface{}) {
	msg := fmt.Sprint(v...)
	lerr.Output(2, tagE+" "+msg+"\n")
	panic(msg)
}

func Fatal(v ...interface{}) {
	lerr.Output(2, tagE+" "+fmt.Sprint(v...)+"\n")
	os.Exit(1)
}
