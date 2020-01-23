package main

import (
	"fmt"
	"testing"

	"gitlab.com/tslocum/netris/pkg/mino"
)

func TestRenderMatrix(t *testing.T) {
	renderLock.Lock()
	defer renderLock.Unlock()

	for bs := 1; bs <= 3; bs++ {
		bs := bs // Capture

		t.Run(fmt.Sprintf("Size=%d", bs), func(t *testing.T) {
			blockSize = bs

			m, err := mino.NewTestMatrix()
			if err != nil {
				t.Error(err)
			}

			m.AddTestBlocks()

			mx := []*mino.Matrix{m}

			renderMatrixes(mx)
		})
	}

	blockSize = 1
}

func BenchmarkRenderMatrix(b *testing.B) {
	renderLock.Lock()
	defer renderLock.Unlock()

	for bs := 1; bs <= 3; bs++ {
		bs := bs // Capture

		b.Run(fmt.Sprintf("Size=%d", bs), func(b *testing.B) {
			blockSize = bs

			m, err := mino.NewTestMatrix()
			if err != nil {
				b.Error(err)
			}

			m.AddTestBlocks()

			mx := []*mino.Matrix{m}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				renderMatrixes(mx)
			}
		})
	}

	blockSize = 1
}
