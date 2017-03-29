package plog

import (
	"log"
	"os"

	"github.com/mattn/go-colorable"
)

var (
	Out = os.Stdout
)

func New() *log.Logger {
	if Out == os.Stdout {
		return log.New(colorable.NewColorableStdout(), "", log.Ltime)
	} else {
		return log.New(Out, "", log.Ltime)
	}
}
