// Package ui owns the program's process streams.
//
// appspec/07 "Output streams" makes the stream each message lands on a
// contract rather than a cosmetic detail, so every write goes through an IO
// value instead of reaching for os.Stdout/os.Stderr directly. Colored output
// is layered on top of this in a later change; the routing is the part that is
// load-bearing today.
package ui

import (
	"fmt"
	"io"
	"os"
)

// IO is the set of streams a run reads from and writes to.
type IO struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

// Std returns the IO bound to the real process streams.
func Std() IO {
	return IO{In: os.Stdin, Out: os.Stdout, Err: os.Stderr}
}

// Print writes a line to stdout.
func (io IO) Print(s string) { fmt.Fprintln(io.Out, s) }

// Printf writes a formatted line to stdout.
func (io IO) Printf(format string, args ...any) { fmt.Fprintf(io.Out, format+"\n", args...) }

// Error writes a line to stderr.
func (io IO) Error(s string) { fmt.Fprintln(io.Err, s) }

// Errorf writes a formatted line to stderr.
func (io IO) Errorf(format string, args ...any) { fmt.Fprintf(io.Err, format+"\n", args...) }
