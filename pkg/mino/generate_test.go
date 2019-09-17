package mino

import (
	"log"
	"testing"
)

func TestGenerate(t *testing.T) {
	var (
		minos []Mino
		err   error
	)
	for _, d := range minoTestData {
		minos, err = Generate(d.Rank)
		if err != nil {
			t.Errorf("failed to generate minos for rank %d: %s", d.Rank, err)
		}

		if len(minos) != len(d.Minos) {
			t.Errorf("failed to generate minos for rank %d: expected to generate %d minos, got %d", d.Rank, len(d.Minos), len(minos))
		}

		for i, ex := range d.Minos {
			found := 0
			for _, m := range minos {
				log.Println(d.Rank)
				log.Println(m.String())
				log.Println("\n" + m.Render())
				if m.String() == ex {
					found++
				}
			}
			if found != 1 {
				t.Errorf("failed to generate minos for rank %d mino %d: expected to generate 1 mino %s - got %d", d.Rank, i, ex, found)
			}
		}
	}
}

func BenchmarkGenerate(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	var (
		minos []Mino
		err   error
	)
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		resetCachedMinos()
		b.StartTimer()

		for _, d := range minoTestData {
			minos, err = Generate(d.Rank)
			if err != nil {
				b.Errorf("failed to generate minos: %s", err)
			}

			if len(minos) != len(d.Minos) {
				b.Errorf("failed to generate minos for rank %d: expected to generate %d minos, got %d", d.Rank, len(d.Minos), len(minos))
			}
		}
	}
}
