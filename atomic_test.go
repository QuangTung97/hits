package hits

import (
	"sync/atomic"
	"testing"
)

func BenchmarkAtomicStore(b *testing.B) {
	var value uint64

	run := func(b *testing.B, count int) {
		for n := 0; n < b.N; n++ {
			for i := 0; i < count; i++ {
				atomic.StoreUint64(&value, 123)
			}
		}
	}

	b.Run("100", func(b *testing.B) { run(b, 100) })
	b.Run("1000", func(b *testing.B) { run(b, 1000) })
	b.Run("10000", func(b *testing.B) { run(b, 10000) })
}

var result uint64

func BenchmarkAtomicLoad(b *testing.B) {
	var value uint64 = 234

	run := func(b *testing.B, count int) {
		for n := 0; n < b.N; n++ {
			for i := 0; i < count; i++ {
				result = atomic.LoadUint64(&value)
			}
		}
	}

	b.Run("100", func(b *testing.B) { run(b, 100) })
	b.Run("1000", func(b *testing.B) { run(b, 1000) })
	b.Run("10000", func(b *testing.B) { run(b, 10000) })
}
