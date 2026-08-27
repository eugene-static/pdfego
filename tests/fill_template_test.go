package tests

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"
)

func BenchmarkUPD_FillTemplate(b *testing.B) {
	core, err := initCore()
	if err != nil {
		b.Fatal(err)
	}

	view := newView(10000)
	template := newUpdTemplate(core, view)

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_, err = template.fill()
		if err != nil {
			b.Fatal(err)
		}
	}

	b.StopTimer()
}

func TestUPD_FillTemplate(t *testing.T) {
	core, err := initCore()
	if err != nil {
		t.Fatal(err)
	}

	view := newView(1)
	template := newUpdTemplate(core, view)

	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	wg := new(sync.WaitGroup)

	for i := range 1 {
		wg.Add(1)

		go func(index int, tt *testing.T) {
			defer wg.Done()

			_time := time.Now()

			bytes, err := template.fill()
			if err != nil {
				t.Error(err)

				return
			}

			tt.Logf("Длительность формирования одной итерации: %v\n", time.Since(_time))

			name := fmt.Sprintf("./output/output_%d.pdf", index)

			output, err := os.Create(name)
			if err != nil {
				t.Error(err)

				return
			}

			defer output.Close()

			_, err = output.Write(bytes)
			if err != nil {
				t.Error(err)

				return
			}

		}(i, t)
	}

	wg.Wait()

	if t.Failed() {
		return
	}

	var memAfter runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memAfter)

	totalBytesAllocated := memAfter.TotalAlloc - memBefore.TotalAlloc
	totalObjectsAllocated := memAfter.Mallocs - memBefore.Mallocs

	t.Logf("Выделено памяти (total_alloc): %.2fМБ", float64(totalBytesAllocated)/(1<<20))
	t.Logf("Выделено памяти в куче (heap_in_use): %.2fМБ", float64(memAfter.HeapInuse)/(1<<20))
	t.Logf("Количество аллокаций (malloc): %d", totalObjectsAllocated)
}
