package color

import (
	"sync"

	"github.com/mgutz/ansi"
)

var (
	A Colorer
)

type Colorer struct {
	k    int
	seen map[string]int
	prev int
	lock sync.Mutex
}

// Return an ansi colored string
func (a *Colorer) Color(x string) string {
	a.lock.Lock()
	defer a.lock.Unlock()

	var format string

	if a.seen == nil {
		a.seen = make(map[string]int)
	}

	var i int
	var ok bool
	if i, ok = a.seen[x]; ok {
	} else {
		// Choose a color unlike the last color
		for {
			i = a.k
			a.k = (a.k + 1) % 5
			if a.prev == 0 || i+1 != a.prev {
				break
			}
		}
		a.seen[x] = i
	}
	a.prev = i + 1

	switch i {
	case 0:
		format = "red:white"
	case 1:
		format = "green:white"
	case 2:
		format = "blue:white"
	case 3:
		format = "cyan:white"
	case 4:
		format = "magenta:white"
	default:
	}
	return ansi.Color(x, format)
}
