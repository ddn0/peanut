package pdo

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

func TestUnique(t *testing.T) {
	N := 100
	type thing struct{}
	var lock sync.Mutex
	var count int

	var items []interface{}
	for i := 0; i < N; i += 1 {
		items = append(items, &thing{})
	}
	if err := DoAll(DoAllOpt{
		Func: func(context.Context, interface{}) error {
			lock.Lock()
			defer lock.Unlock()
			count += 1
			return nil
		},
		Items: items,
	}); err != nil {
		t.Error(err)
	}

	if count != N {
		t.Errorf("%d != %d", count, N)
	}
}

func TestUniqueWithGen(t *testing.T) {
	N := 100
	M := 2
	type thing struct{}
	var lock sync.Mutex
	var count int

	var items []interface{}
	for i := 0; i < N; i += 1 {
		items = append(items, &thing{})
	}
	if err := DoAll(DoAllOpt{
		GenFunc: func(context.Context, interface{}) (items []interface{}, err error) {
			for i := 0; i < M; i += 1 {
				items = append(items, &thing{})
			}
			return
		},
		Func: func(context.Context, interface{}) error {
			lock.Lock()
			defer lock.Unlock()
			count += 1
			return nil
		},
		Items: items,
	}); err != nil {
		t.Error(err)
	}

	if count != N*M {
		t.Errorf("%d != %d", count, N*M)
	}
}

func TestError(t *testing.T) {
	N := 100
	type thing struct{}
	var lock sync.Mutex
	var count int

	bad := fmt.Errorf("bad")
	var items []interface{}
	for i := 0; i < N; i += 1 {
		items = append(items, &thing{})
	}
	if err := DoAll(DoAllOpt{
		Func: func(context.Context, interface{}) error {
			lock.Lock()
			defer lock.Unlock()
			count += 1
			if count == N-1 {
				return bad
			}
			return nil
		},
		Items: items,
	}); err != bad {
		t.Error("expecting error %v but found %v", bad, err)
	}
}
