package util

import (
	"context"
	"runtime/debug"
	"sync"

	"github.com/58kg/logs"
)

func AllDone(ctx context.Context, fs ...func()) {
	var wg sync.WaitGroup
	for _, f := range fs {
		wg.Add(1)
		ff := f
		task := func() {
			defer func() {
				wg.Done()
				if err := recover(); err != nil {
					logs.CtxCritical(ctx, "task panic, err:%v, stack:\n%s", err, debug.Stack())
				}
			}()
			ff()
		}
		go task()
	}
	wg.Wait()
}
