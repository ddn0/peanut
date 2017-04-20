// io.Writer interface around log.Logger that outputs each line of output as a
// separate Logger.Println() call. This breaks the atomicity guarantee of
// Logger methods, but gives us prettier output.
package logwriter

import (
	"bytes"
	"fmt"
	"log"
	"sync"
)

type Options struct {
	Prepend string
}

type LogWriter struct {
	l    *log.Logger
	opts Options
	buf  bytes.Buffer
	lock sync.Mutex
}

func (a *LogWriter) println(b []byte) {
	if len(a.opts.Prepend) > 0 {
		a.l.Println(a.opts.Prepend, string(b))
	} else {
		a.l.Println(string(b))
	}
}

func (a *LogWriter) output(b []byte) int {
	if len(b) == 0 {
		return 0
	}
	prevI := 0
	i := bytes.IndexByte(b[prevI:], '\n')
	for 0 <= i && prevI+i < len(b) {
		a.println(b[prevI : prevI+i])
		prevI = prevI + i + 1
		if prevI >= len(b) {
			break
		}
		i = bytes.IndexByte(b[prevI:], '\n')
	}
	return prevI
}

func (a *LogWriter) Write(p []byte) (n int, err error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.buf.Write(p)
	written := a.output(a.buf.Bytes())
	a.buf.Next(written)
	return len(p), nil
}

func (a *LogWriter) Printf(format string, args ...interface{}) error {
	s := fmt.Sprintf(format, args...)
	_, err := a.Write([]byte(s))
	return err
}

func (a *LogWriter) Flush() error {
	a.lock.Lock()
	defer a.lock.Unlock()
	written := a.output(a.buf.Bytes())
	a.buf.Next(written)
	b := a.buf.Bytes()
	if len(b) > 0 {
		a.println(b)
	}
	return nil
}

func New(l *log.Logger, opts *Options) *LogWriter {
	if opts == nil {
		opts = &Options{}
	}
	return &LogWriter{l: l, opts: *opts}
}
