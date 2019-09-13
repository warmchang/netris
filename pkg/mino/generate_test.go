package mino

import "testing"

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

		if len(minos) != d.Minos {
			t.Errorf("failed to generate minos for rank %d: expected to generate %d minos, got %d", d.Rank, d.Minos, len(minos))
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

			if len(minos) != d.Minos {
				b.Errorf("failed to generate minos for rank %d: expected to generate %d minos, got %d", d.Rank, d.Minos, len(minos))
			}
		}
	}
}
