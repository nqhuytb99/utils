package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

var (
	dataSmall         = (`{"id":12125925,"ids":[-2147483648,2147483647],"title":"未来简史-从智人到智神","titles":["hello","world"],"price":40.8,"prices":[-0.1,0.1],"hot":true,"hots":[true,true,true],"author":{"name":"json","age":99,"male":true},"authors":[{"name":"json","age":99,"male":true},{"name":"json","age":99,"male":true},{"name":"json","age":99,"male":true}],"weights":[]}`)
	insertConcurrency = 2
)

func BenchmarkCache(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c := New[string](WithCapacity(100 * 1000))
		wg := new(sync.WaitGroup)
		wg.Add(insertConcurrency)

		for i := 0; i < insertConcurrency; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < 10*1000; j++ {
					key := uuid.NewString()
					value := dataSmall + fmt.Sprint(j)
					c.Set(key, value, 1*time.Second)

					data, exist := c.Get(key)
					if !exist || data != value {
						b.Fail()
					}
				}
			}()
		}

		wg.Wait()
	}
}
