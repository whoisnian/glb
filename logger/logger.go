// Package logger implements a simple logger for local debug.
//   - [I]: tagI for info log
//   - [W]: tagW for warn log
//   - [E]: tagE for error log
//   - [D]: tagD for debug log
//   - [R]: tagR for request log
//
// By default, colorful flag will be enable if standard output is connected to a terminal.
package logger

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/whoisnian/glb/ansi"
	"golang.org/x/term"
)

const labelI, labelW, labelE, labelD, labelR string = "[I]", "[W]", "[E]", "[D]", "[R]"

var debug, colorful bool = false, false
var tagI, tagW, tagE, tagD, tagR string = labelI, labelW, labelE, labelD, labelR
var lout, lerr *log.Logger

func init() {
	lout = log.New(os.Stdout, "", log.LstdFlags)
	lerr = log.New(os.Stderr, "", log.LstdFlags)
	if term.IsTerminal(int(os.Stdout.Fd())) {
		SetColorful(true)
	}
}

// SetOutput redirects stdout/stderr of the logger.
func SetOutput(stdout, stderr io.Writer) {
	if stdout != nil {
		lout.SetOutput(stdout)
	}
	if stderr != nil {
		lerr.SetOutput(stderr)
	}
}

// SetDebug set debug flag.
func SetDebug(enable bool) {
	debug = enable
}

// SetColorful set colorful flag.
func SetColorful(enable bool) {
	colorful = enable
	if enable {
		tagI = ansi.GreenFG + labelI + ansi.Reset
		tagW = ansi.YellowFG + labelW + ansi.Reset
		tagE = ansi.RedFG + labelE + ansi.Reset
		tagD = ansi.MagentaFG + labelD + ansi.Reset
		tagR = ansi.BlueFG + labelR + ansi.Reset
	} else {
		tagI, tagW, tagE, tagD, tagR = labelI, labelW, labelE, labelD, labelR
	}
}

// IsDebug return debug flag.
func IsDebug() bool {
	return debug
}

// IsColorful return colorful flag.
func IsColorful() bool {
	return colorful
}

// Info writes info log to stdout with tagI.
func Info(v ...interface{}) {
	lout.Output(2, tagI+" "+fmt.Sprint(v...)+"\n")
}

// Warn writes warn log to stdout with tagW.
func Warn(v ...interface{}) {
	lout.Output(2, tagW+" "+fmt.Sprint(v...)+"\n")
}

// Error writes error log to stderr with tagE.
func Error(v ...interface{}) {
	lerr.Output(2, tagE+" "+fmt.Sprint(v...)+"\n")
}

// Debug writes debug log to stdout with tagD.
//
// If parameters need heavy calculation, should wrap them with a conditional block. Example:
//
//	if logger.IsDebug() {
//	    diffs := calcDifference(A, B)
//	    logger.Debug("compare A to B:", diffs)
//	}
func Debug(v ...interface{}) {
	if !debug {
		return
	}
	lout.Output(2, tagD+" "+fmt.Sprint(v...)+"\n")
}

// Panic is equivalent to logger.Error() followed by a call to panic().
func Panic(v ...interface{}) {
	msg := fmt.Sprint(v...)
	lerr.Output(2, tagE+" "+msg+"\n")
	panic(msg)
}

// Fatal is equivalent to logger.Error() followed by a call to os.Exit(1).
func Fatal(v ...interface{}) {
	lerr.Output(2, tagE+" "+fmt.Sprint(v...)+"\n")
	os.Exit(1)
}
