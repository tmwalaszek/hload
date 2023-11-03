package cliio

import (
	"io"
	"os"
)

type IO struct {
	In io.Reader

	Out io.Writer
	Err io.Writer
}

func NewStdIO() IO {
	return IO{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	}
}
