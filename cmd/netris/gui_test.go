package main

import (
	"testing"

	"git.sr.ht/~tslocum/netris/pkg/mino"
)

func TestRenderMatrix(t *testing.T) {
	renderLock.Lock()
	defer renderLock.Unlock()

	blockSize = 1

	m, err := mino.NewTestMatrix()
	if err != nil {
		t.Error(err)
	}

	m.AddTestBlocks()

	mx := []*mino.Matrix{m}

	renderMatrixes(mx)
}

func BenchmarkRenderStandardMatrix(b *testing.B) {
	renderLock.Lock()
	defer renderLock.Unlock()

	blockSize = 1

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
}

func BenchmarkRenderLargeMatrix(b *testing.B) {
	renderLock.Lock()
	defer renderLock.Unlock()

	blockSize = 2

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

	blockSize = 1
}
