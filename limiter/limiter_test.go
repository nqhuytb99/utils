package limiter

import (
	"testing"
	"time"
)

const (
	reqPerSec      = 10
	concurrentUser = 10 * 1000
)

func BenchmarkLimiter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		l := NewLimiter(1*time.Second, reqPerSec*concurrentUser)
		_ = l
		for j := 0; j < 200*1000; j++ {
			l.Inc()
		}
	}
}

func TestLimiter(t *testing.T) {
	l := NewLimiter(10*time.Second, reqPerSec*concurrentUser)
	success, fail := 0, 0

	for j := 0; j < 3*reqPerSec*concurrentUser; j++ {
		if l.Inc() {
			success++
		} else {
			fail++
		}
	}

	if success != reqPerSec*concurrentUser || fail != 2*reqPerSec*concurrentUser {
		t.Fail()
	}
}
