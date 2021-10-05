package util

import (
	"context"
	"runtime/debug"

	"github.com/58kg/logs"
)

func Do(ctx context.Context, task func()) <-chan struct{} {
	doChan := make(chan struct{})
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logs.CtxCritical(ctx, "task panic, err:%v, stack:\n%s", err, debug.Stack())
			}
			close(doChan)
		}()
		task()
	}()
	return doChan
}
