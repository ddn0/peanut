// Parallel DoAll
package pdo

import (
	"context"
	"sync"
	"time"
)

type Func func(context.Context, interface{}) error

type GenFunc func(context.Context, interface{}) ([]interface{}, error)

type DoAllOpt struct {
	GenFunc       GenFunc       // If not nil, make a two-stage pipeline Func(GenFunc(Item))
	Func          Func          // Func(Item)
	Items         []interface{} // Iteration space
	Timeout       time.Duration // Maximum duration of any function
	MaxConcurrent int           // Maximum number of concurrent threads per stage minus 1
}

func DoAll(opt DoAllOpt) error {
	timeout := opt.Timeout
	if timeout == 0 {
		timeout = 1 * time.Second
	}

	doOne := func(pctx context.Context, item interface{}) error {
		ctx, cancel := context.WithDeadline(pctx, time.Now().Add(timeout))
		defer cancel()
		return opt.Func(ctx, item)
	}

	doGen := func(pctx context.Context, item interface{}) ([]interface{}, error) {
		ctx, cancel := context.WithDeadline(pctx, time.Now().Add(timeout))
		defer cancel()
		return opt.GenFunc(ctx, item)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var next chan interface{}
	items := make(chan interface{})
	done := make(chan bool)
	errs := make(chan error)
	defer close(done)

	// Distributor
	go func() {
		defer close(items)
		for _, item := range opt.Items {
			select {
			case items <- item:
			case <-done:
				return
			}
		}
	}()

	var errsWg sync.WaitGroup
	num := opt.MaxConcurrent + 1
	if opt.GenFunc != nil {
		errsWg.Add(num)
	}
	errsWg.Add(num)

	if opt.GenFunc != nil {
		next = make(chan interface{})

		var wg sync.WaitGroup
		wg.Add(num)

		for idx := 0; idx < num; idx += 1 {
			go func() {
				defer errsWg.Done()
				defer wg.Done()
				for {
					select {
					case <-done:
						return
					case item := <-items:
						if item == nil {
							return
						}
						ns, err := doGen(ctx, item)
						errs <- err
						if err != nil {
							return
						}
						for _, n := range ns {
							next <- n
						}
					}
				}
			}()
		}

		go func() {
			wg.Wait()
			close(next)
		}()
	} else {
		next = items
	}

	// Worker
	for idx := 0; idx < num; idx += 1 {
		go func() {
			defer errsWg.Done()
			for {
				select {
				case <-done:
					return
				case item := <-next:
					if item == nil {
						return
					}
					err := doOne(ctx, item)
					errs <- err
					if err != nil {
						return
					}
				}
			}
		}()
	}

	// Waiter
	go func() {
		errsWg.Wait()
		close(errs)
	}()

	// Collect results
	var err error
	for e := range errs {
		if e != nil && err == nil {
			err = e
		}
	}
	return err
}
