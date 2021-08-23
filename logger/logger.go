// Package logger implements a simple logger for local debug.
package logger

import (
	"fmt"
	"log"
	"os"

	"github.com/whoisnian/glb/ansi"
)

const labelI, labelW, labelE, labelD, labelR string = "[I]", "[W]", "[E]", "[D]", "[R]"

var debug bool = false
var lout, lerr *log.Logger
var tagI, tagW, tagE, tagD, tagR string = labelI, labelW, labelE, labelD, labelR

func init() {
	lout = log.New(os.Stdout, "", log.LstdFlags)
	lerr = log.New(os.Stderr, "", log.LstdFlags)
}

func SetDebug(value bool) {
	debug = value
	SetColorful(debug)
}

func SetColorful(value bool) {
	if value {
		tagI = ansi.Green + labelI + ansi.Reset
		tagW = ansi.Yellow + labelW + ansi.Reset
		tagE = ansi.Red + labelE + ansi.Reset
		tagD = ansi.Magenta + labelD + ansi.Reset
		tagR = ansi.Blue + labelR + ansi.Reset
	} else {
		tagI, tagW, tagE, tagD, tagR = labelI, labelW, labelE, labelD, labelR
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

func Panic(v ...interface{}) {
	msg := fmt.Sprint(v...)
	lerr.Output(2, tagE+" "+msg+"\n")
	panic(msg)
}

func Fatal(v ...interface{}) {
	lerr.Output(2, tagE+" "+fmt.Sprint(v...)+"\n")
	os.Exit(1)
}
