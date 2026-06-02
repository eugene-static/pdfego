package upd

import (
	"os"
	"runtime"
	"testing"
	"time"
)

func BenchmarkUPD_FillTemplate(b *testing.B) {
	upd := NewUPD(10)

	template, err := prepareTemplate()

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_, err = upd.fillTemplate(template)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.StopTimer()
}

func TestUPD_FillTemplate(t *testing.T) {
	upd := NewUPD(5000)

	template, err := prepareTemplate()
	if err != nil {
		t.Fatal(err)
	}

	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	_time := time.Now()

	bytes, err := upd.fillTemplate(template)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Длительность формирования одной итерации: %v\n", time.Since(_time))

	var memAfter runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memAfter)

	totalBytesAllocated := memAfter.TotalAlloc - memBefore.TotalAlloc
	totalObjectsAllocated := memAfter.Mallocs - memBefore.Mallocs

	t.Logf("Выделено памяти за одну итерацию (total_alloc): %.2fМБ", float64(totalBytesAllocated)/(1<<20))
	t.Logf("Выделено памяти в куче (heap_in_use): %.2fМБ", float64(memAfter.HeapInuse)/(1<<20))
	t.Logf("Количество аллокаций за одну итерацию (malloc): %d", totalObjectsAllocated)

	output, err := os.Create("output.pdf")
	if err != nil {
		t.Fatal(err)
	}

	defer output.Close()

	outputText, err := os.Create("output_text.txt")
	if err != nil {
		t.Fatal(err)
	}

	defer outputText.Close()

	_, err = output.Write(bytes)
	if err != nil {
		t.Fatal(err)
	}

	_, err = outputText.Write(bytes)
	if err != nil {
		t.Fatal(err)
	}
}
