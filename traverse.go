package util

import (
	"fmt"
	"github.com/58kg/gpool"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

type Node interface {
	Id() string       // 结点唯一标识
	Task()            // 结点执行的函数
	Children() []Node // 所有的子结点
}

func Traverse(root Node, multiRoutine bool) {
	if !multiRoutine {
		(&singleDFS{
			visit: map[string]struct{}{root.Id(): {}},
		}).dfs(root)
		return
	}

	h := &multiDFS{}
	h.visit.Store(root.Id(), struct{}{})
	h.gp, _ = gpool.NewPool(gpool.Config{
		MaxTaskWorkerCount: gpool.MaxTaskWorkerCount,
		KeepaliveTime:      time.Millisecond * 10,
		HandleTaskPanic: func(err interface{}) {
			_, _ = fmt.Fprintf(os.Stderr, "task panic, err:%s, stack:\n%s", err, debug.Stack())
		},
	})
	h.dfs(root)
	h.wg.Wait()
	_ = h.gp.Stop()
}

type singleDFS struct {
	visit map[string]struct{}
}

func (h *singleDFS) dfs(root Node) {
	root.Task()
	for _, ch := range root.Children() {
		if _, inMap := h.visit[ch.Id()]; inMap {
			continue
		}
		h.visit[ch.Id()] = struct{}{}
		h.dfs(ch)
	}
}

type multiDFS struct {
	visit sync.Map
	gp    *gpool.Pool
	wg    sync.WaitGroup
}

func (h *multiDFS) dfs(root Node) {
	root.Task()
	for _, ch := range root.Children() {
		if _, loaded := h.visit.LoadOrStore(ch.Id(), struct{}{}); loaded {
			continue
		}

		chCopy := ch
		h.wg.Add(1)
		task := func() {
			defer h.wg.Done()
			h.dfs(chCopy)
		}

		if succ, _ := h.gp.AddTask(task, nil, false); succ {
			continue
		}

		task()
	}
}
