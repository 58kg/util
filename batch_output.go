package util

import (
	"fmt"
	"sync"
	"time"
)

type BatchOutput struct {
	batchSize         uint64
	flushTimeInterval time.Duration
	outputChan        chan []interface{}
	closeSignalChan   chan struct{}
	elements          []interface{}
	cancel            func()
	lock              sync.Mutex
}

type BatchOutputStatus struct {
	BatchSize         uint64
	FlushTimeInterval time.Duration
}

// BatchOutput用于将输入元素聚合后输出, 主要参数为BatchSize和FlushTimeInterval, 含义如下:
//    BatchSize: 聚合的batch大小
//    FlushTimeInterval: 超过FlushTimeInterval时间内输入元素的个数未达到BatchSize时, 将当前剩余所有元素强制聚合后输出
func NewBatchOutput(batchSize uint64, flushTimeInterval time.Duration) *BatchOutput {
	if batchSize == 0 {
		batchSize = 1
	}
	ret := &BatchOutput{
		batchSize:         batchSize,
		flushTimeInterval: flushTimeInterval,
		outputChan:        make(chan []interface{}),
		closeSignalChan:   make(chan struct{}),
	}
	// 启动循环监控elements的协程
	ret.cancel = CycleRun(func() {
		ret.flushElements(true)
	}, nil, ret.flushTimeInterval, nil)
	return ret
}

func (b *BatchOutput) PushElement(x interface{}) error {
	if !b.isInited() {
		return batchOutputNotInitError
	}
	if b.isStopped() {
		return batchOutputStoppedError
	}
	var cancelFunc func()
	b.lock.Lock()
	defer func() {
		b.lock.Unlock()
		if cancelFunc != nil {
			cancelFunc()
		}
	}()
	b.elements = append(b.elements, x)
	if uint64(len(b.elements)) >= b.batchSize {
		select {
		case b.outputChan <- b.elements:
			b.elements = nil
			cancelFunc = b.cancel
			b.cancel = CycleRun(func() {
				b.flushElements(true)
			}, nil, b.flushTimeInterval, nil)
		case <-b.closeSignalChan:
			return batchOutputStoppedError
		}
	}
	return nil
}

func (b *BatchOutput) ChangeBatchSizeOrFlushTimeInterval(newBatchSize *uint64, newFlushTimeInterval *time.Duration) error {
	if !b.isInited() {
		return batchOutputNotInitError
	}
	if b.isStopped() {
		return batchOutputStoppedError
	}
	if newBatchSize != nil && *newBatchSize == 0 {
		t := uint64(1)
		newBatchSize = &t
	}
	var cancelFunc func()
	b.lock.Lock()
	defer func() {
		b.lock.Unlock()
		if cancelFunc != nil {
			cancelFunc()
		}
	}()
	var changed bool
	if newBatchSize != nil && b.batchSize != *newBatchSize {
		b.batchSize = *newBatchSize
		changed = true
	}
	if newFlushTimeInterval != nil && *newFlushTimeInterval != b.flushTimeInterval {
		b.flushTimeInterval = *newFlushTimeInterval
		changed = true
		cancelFunc = b.cancel
		b.cancel = CycleRun(func() {
			b.flushElements(true)
		}, nil, b.flushTimeInterval, nil)
	}
	if changed {
		b.flushElements(false)
	}
	return nil
}

func (b *BatchOutput) GetBatchElements(withBlock bool) ([]interface{}, error) {
	if !b.isInited() {
		return nil, batchOutputNotInitError
	}
	if b.isStopped() {
		return nil, batchOutputStoppedError
	}
	if !withBlock {
		select {
		case elements := <-b.outputChan:
			return elements, nil
		default:
			return nil, nil
		}
	}
	select {
	case elements := <-b.outputChan:
		return elements, nil
	case <-b.closeSignalChan:
		return nil, batchOutputStoppedError
	}
}

func (b *BatchOutput) GetStatus() (BatchOutputStatus, error) {
	if !b.isInited() {
		return BatchOutputStatus{}, batchOutputNotInitError
	}
	if b.isStopped() {
		return BatchOutputStatus{}, batchOutputStoppedError
	}
	b.lock.Lock()
	defer b.lock.Unlock()
	return BatchOutputStatus{
		BatchSize:         b.batchSize,
		FlushTimeInterval: b.flushTimeInterval,
	}, nil
}

func (b *BatchOutput) Stop() {
	close(b.closeSignalChan)
}

func (b *BatchOutput) flushElements(withLock bool) {
	if withLock {
		b.lock.Lock()
		defer b.lock.Unlock()
	}
	if len(b.elements) > 0 {
		select {
		case b.outputChan <- b.elements:
			b.elements = nil
		case <-b.closeSignalChan:
		}
	}
}

func (b *BatchOutput) isStopped() bool {
	select {
	case <-b.closeSignalChan:
		return true
	default:
		return false
	}
}

func (b *BatchOutput) isInited() bool {
	return b.batchSize > 0
}

var (
	batchOutputStoppedError = fmt.Errorf("error: BatchOutput is stopped")
	batchOutputNotInitError = fmt.Errorf("error: BatchOutput is not inited")
)
