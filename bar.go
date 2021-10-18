package util

import (
	"fmt"
	"os"
	"sync"
	"unicode/utf8"
)

const (
	Highlight                 = 1
	FrontGreen                = 32
	BackBlack                 = 40
	BasicCompletedIcon   rune = '█'
	BasicUnCompletedIcon byte = ' '
)

type Bar interface {
	// 更新成功时返回true
	Update(cur uint64) bool
}

type bar struct {
	percent float64 //百分比
	curVal  uint64  //当前进度位置
	total   uint64  //总进度
	bar     []byte  //进度条
	left    int
	right   int
	sync.Mutex
}

// NewBarWithConfig 通过自定义配置对象创建进度条对象
func NewBar(total uint64) Bar {
	b := &bar{
		total: total,
		bar:   make([]byte, 201),
		right: 50,
	}

	for i := 0; i < 50; i++ {
		b.bar[i] = BasicUnCompletedIcon
	}
	return b
}

// 获取当前状态百分比
func (b *bar) getPercent() float64 {
	return float64(b.curVal) / float64(b.total) * 100
}

// Update 执行一次记录
func (b *bar) Update(cur uint64) bool {
	b.Lock()
	defer b.Unlock()
	if cur <= b.curVal {
		return false
	}

	if cur > b.total {
		cur = b.total
	}

	b.curVal = cur
	latestPercent := b.getPercent()

	var t = int(latestPercent) - int(b.percent)
	if b.percent != latestPercent && t%2 == 0 {
		b.percent = latestPercent
		t /= 2
		var start, n, i int
		start = b.right
		for i = 0; i < t; i++ {
			n = utf8.EncodeRune(b.bar[b.left:b.left+utf8.UTFMax], BasicCompletedIcon)
			b.left += n
			b.right += n - 1
		}
		b.excursion(start, b.right+n-1)
	}
	// 将bar的[0:右边界]打印出来
	doPrint(b.bar[:b.right], b.percent, int64(b.curVal), int64(b.total))
	return true
}

func (b *bar) excursion(start, end int) {
	for i := start; i < end; i++ {
		if int(b.bar[i]) < 100 {
			b.bar[i] = BasicUnCompletedIcon //b.config.unCompletedIcon
		}
	}
}

const format = "\rProgress:[\u001B[%d;%d;%dm%s\u001B[0m] %3.2f%% Completed:[%3d] Total:[%3d]"

func doPrint(str []byte, percent float64, currPos, total int64) {
	_, _ = fmt.Fprintf(os.Stdout, format, Highlight, BackBlack, FrontGreen, string(str), percent, currPos, total)
}
