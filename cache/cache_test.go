package cache

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkCache(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c := New[string]()
		for j := 0; j < 10*1000; j++ {
			c.Set(fmt.Sprint(j), fmt.Sprint(j), 1*time.Second)

			data, exist := c.Get(fmt.Sprint(j))
			if !exist || data != fmt.Sprint(j) {
				b.Fail()
			}
		}
	}
}
