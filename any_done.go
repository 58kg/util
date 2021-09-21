package util

import (
	"context"
	"runtime/debug"
	"sync/atomic"

	"github.com/gogokit/logs"
)

// 并发执行fs中的所有函数, 且在fs中某个函数最先返回(res0, true)时返回(res0, true), 如果fs中所有函数均返回(res, false), 则在fs中所有函数执行完之后返回(nil, false)
func AnyDone(ctx context.Context, fs ...func() (res interface{}, needReturn bool)) (res interface{}, withNeedReturn bool) {
	if len(fs) == 0 {
		// len(fs)为0时, 如果继续下面的流程会导致协程阻塞且永远不能被唤醒
		return nil, false
	}

	resultSignalChan := make(chan interface{}, len(fs))
	allDoneSignalChan := make(chan struct{})
	doneCount := int32(len(fs))
	resNotNilCount := int32(0)

	for _, f := range fs {
		go func(g func() (interface{}, bool)) {
			defer func() {
				if atomic.AddInt32(&doneCount, -1) == 0 && atomic.LoadInt32(&resNotNilCount) == 0 {
					close(allDoneSignalChan)
				}
				if err := recover(); err != nil {
					logs.CtxCritical(ctx, "task panic, err:%v, stack:\n%s", err, debug.Stack())
				}
			}()
			if res, needReturn := g(); needReturn {
				resultSignalChan <- res
				atomic.StoreInt32(&resNotNilCount, 1)
			}
		}(f)
	}
	select {
	case ret := <-resultSignalChan:
		return ret, true
	case <-allDoneSignalChan:
		return nil, false
	}
}
