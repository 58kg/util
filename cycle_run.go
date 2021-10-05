package util

import (
	"time"
)

//　以周期timeInterval执行函数f, 调用cancel()可以停止周期执行任务, 多次调用cancel会panic
func CycleRun(f func(), handleFPanic func(interface{}), timeInterval time.Duration, maxRunTimes *uint64) (cancel func()) {
	loopFinishSignalChan := make(chan struct{})
	closeSignalChan := make(chan struct{})
	cancel = func() {
		close(closeSignalChan)
		<-loopFinishSignalChan
	}
	go func() {
		defer close(loopFinishSignalChan)
		for loopTimes := uint64(0); maxRunTimes == nil || (loopTimes < *maxRunTimes); loopTimes++ {
			timer := time.NewTimer(timeInterval)
			select {
			case <-timer.C:
			case <-closeSignalChan:
				timer.Stop()
				return
			}
			func() {
				defer func() {
					if err := recover(); err != nil && handleFPanic != nil {
						handleFPanic(err)
					}
				}()
				f()
			}()
		}
	}()
	return
}
