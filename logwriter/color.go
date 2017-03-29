package logwriter

import (
	"github.com/ddn0/peanut/color"
	"github.com/ddn0/peanut/plog"
)

// Create new logwriter with prefix in some color
func NewColorWriter(p string) *LogWriter {
	if len(p) == 0 {
		return New(plog.New(), &Options{})
	} else {
		return New(plog.New(), &Options{Prepend: color.A.Color(p)})
	}
}
