package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

// BenchmarkStringOperations benchmarks string operations
func BenchmarkStringOperations(b *testing.B) {
	b.Run("StringConcatenation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := "Hello" + " " + "World" + "!"
			_ = result
		}
	})

	b.Run("StringBuilderConcatenation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var builder strings.Builder
			builder.WriteString("Hello")
			builder.WriteString(" ")
			builder.WriteString("World")
			builder.WriteString("!")
			_ = builder.String()
		}
	})

	b.Run("StringSplit", func(b *testing.B) {
		s := "one,two,three,four,five"
		for i := 0; i < b.N; i++ {
			parts := strings.Split(s, ",")
			_ = parts
		}
	})
}

// BenchmarkJSONOperations benchmarks JSON marshaling/unmarshaling
func BenchmarkJSONOperations(b *testing.B) {
	type TestData struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Data []int  `json:"data"`
	}

	testData := TestData{
		ID:   1,
		Name: "test",
		Data: []int{1, 2, 3, 4, 5},
	}

	b.Run("JSONMarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(testData)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	jsonBytes, _ := json.Marshal(testData)
	b.Run("JSONUnmarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var data TestData
			err := json.Unmarshal(jsonBytes, &data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkSliceOperations benchmarks slice operations
func BenchmarkSliceOperations(b *testing.B) {
	b.Run("SliceAppend", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var slice []int
			for j := 0; j < 100; j++ {
				slice = append(slice, j)
			}
		}
	})

	b.Run("SlicePrealloc", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slice := make([]int, 0, 100)
			for j := 0; j < 100; j++ {
				slice = append(slice, j)
			}
		}
	})

	b.Run("SliceIteration", func(b *testing.B) {
		slice := make([]int, 1000)
		for i := range slice {
			slice[i] = i
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sum := 0
			for _, v := range slice {
				sum += v
			}
			_ = sum
		}
	})
}

// BenchmarkMapOperations benchmarks map operations
func BenchmarkMapOperations(b *testing.B) {
	b.Run("MapWrite", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := make(map[string]int)
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key%d", j)
				m[key] = j
			}
		}
	})

	b.Run("MapRead", func(b *testing.B) {
		m := make(map[string]int)
		for j := 0; j < 100; j++ {
			key := fmt.Sprintf("key%d", j)
			m[key] = j
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key%d", j)
				_ = m[key]
			}
		}
	})
}

// BenchmarkConcurrency benchmarks concurrent operations
func BenchmarkConcurrency(b *testing.B) {
	b.Run("Goroutines", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var wg sync.WaitGroup
			for j := 0; j < 10; j++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					time.Sleep(time.Microsecond)
				}()
			}
			wg.Wait()
		}
	})

	b.Run("Channels", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ch := make(chan int, 10)
			go func() {
				for j := 0; j < 10; j++ {
					ch <- j
				}
				close(ch)
			}()

			for range ch {
				// Process values
			}
		}
	})
}

// BenchmarkContextOperations benchmarks context operations
func BenchmarkContextOperations(b *testing.B) {
	b.Run("ContextWithTimeout", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			_ = ctx
			cancel()
		}
	})

	b.Run("ContextWithCancel", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx, cancel := context.WithCancel(context.Background())
			_ = ctx
			cancel()
		}
	})
}

// BenchmarkRandomOperations benchmarks random number generation
func BenchmarkRandomOperations(b *testing.B) {
	b.Run("RandInt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = rand.Int()
		}
	})

	b.Run("RandIntn", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = rand.Intn(100)
		}
	})
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("SmallAllocation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			data := make([]byte, 64)
			_ = data
		}
	})

	b.Run("LargeAllocation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			data := make([]byte, 64*1024)
			_ = data
		}
	})

	b.Run("StructAllocation", func(b *testing.B) {
		type TestStruct struct {
			ID   int
			Name string
			Data []int
		}

		for i := 0; i < b.N; i++ {
			s := &TestStruct{
				ID:   i,
				Name: fmt.Sprintf("test%d", i),
				Data: make([]int, 10),
			}
			_ = s
		}
	})
}
