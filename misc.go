package util

import (
	"math/rand"
	"regexp"
	"time"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

// RetryDo Retry sn表示这是第几次执行f, 至多执行times次f
func Retry(f func(sn int) (end bool), times int, maxSleepTime time.Duration) {
	for i := 1; i <= times; i++ {
		if f(i) || i == times {
			return
		}
		time.Sleep(time.Duration(random.Int63n(int64(maxSleepTime))))
	}
}

var trimWhiteReg = regexp.MustCompile(`(^\s+)|(\s+$)`)

func TrimWhite(s string) string {
	return trimWhiteReg.ReplaceAllString(s, "")
}
