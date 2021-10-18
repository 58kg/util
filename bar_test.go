package util

import (
	"testing"
	"time"
)

func TestBar(t *testing.T) {
	bar := NewBar(1000)
	for i := 1; i <= 1000; i++ {
		time.Sleep(time.Millisecond * 10)
		bar.Update(uint64(i))
	}
}
