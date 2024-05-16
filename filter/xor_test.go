package filter

import (
	"crypto/rand"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/FastFilter/xorfilter"
)

const noTestingKeys = 10 * 1000 * 1000

var keys = make([]string, noTestingKeys)

func init() {
	start := time.Now()
	for i := 0; i < noTestingKeys; i++ {
		randKey := make([]byte, 32)
		rand.Read(randKey)
		keys[i] = fmt.Sprintf("%x", randKey)
	}
	fmt.Println("Time to generate keys: ", time.Since(start))
}

func TestFilter_SaveToFile(t *testing.T) {
	filter := NewFilter()
	filter.BuildFilter(keys)

	type fields struct {
		xorfilter *xorfilter.BinaryFuse8
	}
	type args struct {
		fp string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:   "Test SaveToFile",
			fields: fields(*filter),
			args: args{
				fp: "xorfilter.gob",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Filter{
				xorfilter: tt.fields.xorfilter,
			}
			if err := f.SaveToFile(tt.args.fp); (err != nil) != tt.wantErr {
				t.Errorf("Filter.SaveToFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFilter_LoadFromFile(t *testing.T) {
	filter := NewFilter()

	type fields struct {
		xorfilter *xorfilter.BinaryFuse8
	}
	type args struct {
		fp string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:   "Test LoadFromFile",
			fields: fields(*filter),
			args: args{
				fp: "xorfilter.gob",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Filter{
				xorfilter: tt.fields.xorfilter,
			}
			if err := f.LoadFromFile(tt.args.fp); (err != nil) != tt.wantErr {
				t.Errorf("Filter.LoadFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFilter_Contains(t *testing.T) {
	filter := NewFilter()
	filter.BuildFilter(keys)

	type fields struct {
		xorfilter *xorfilter.BinaryFuse8
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "Test Contains",
			fields: fields(*filter),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Filter{
				xorfilter: tt.fields.xorfilter,
			}

			for i := 0; i < noTestingKeys; i++ {
				if !f.Contains(keys[i]) {
					t.Errorf("Filter.Contains() = %v, want %v", false, true)
				}
			}

			var falsePositive int
			for i := noTestingKeys; i < 2*noTestingKeys; i++ {
				randKey := make([]byte, 16)
				rand.Read(randKey)
				if f.Contains(fmt.Sprintf("%x", randKey)) {
					falsePositive++
				}
			}
			os.WriteFile("rate", []byte(fmt.Sprintf("False positive rate: %f\n", float64(falsePositive)/float64(noTestingKeys)*100)), 0644)
			fmt.Println("False positive rate: ", float64(falsePositive)/float64(noTestingKeys)*100, "%")
		})
	}
}

func BenchmarkContains(b *testing.B) {
	for i := 0; i < b.N; i++ {
		start := time.Now()
		filter := NewFilter()
		filter.BuildFilter(keys)
		fmt.Println("Time to build filter: ", time.Since(start))
		randKey := make([]byte, 16)
		rand.Read(randKey)

		filter.Contains(string(randKey))
	}
}
