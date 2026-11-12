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

func BenchmarkInsertRuneLinhaTipica(b *testing.B) {
	buf, err := Load(bytes.NewReader([]byte("\t\treturn fmt.Errorf(\"falha ao processar item %d: %w\", id, err)\n")))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		end := buf.Insert(Pos{Line: 0, Col: 20}, "x")
		buf.Delete(Pos{Line: 0, Col: 20}, end)
	}
}

func BenchmarkInsertRuneLinhaGigante(b *testing.B) {
	buf, err := Load(bytes.NewReader(bytes.Repeat([]byte("a"), 100_000)))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		end := buf.Insert(Pos{Line: 0, Col: 50_000}, "x")
		buf.Delete(Pos{Line: 0, Col: 50_000}, end)
	}
}
