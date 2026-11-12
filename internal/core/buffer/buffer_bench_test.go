package buffer

import (
	"bytes"
	"fmt"
	"testing"
)

func BenchmarkLoad100kLines(b *testing.B) {
	var src bytes.Buffer
	for i := 0; i < 100_000; i++ {
		fmt.Fprintf(&src, "\tif err := process(ctx, %d); err != nil { return err } // ação\n", i)
	}
	data := src.Bytes()
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	for b.Loop() {
		if _, err := Load(bytes.NewReader(data)); err != nil {
			b.Fatal(err)
		}
	}
}
