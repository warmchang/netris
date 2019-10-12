package main

import (
	"testing"

	"git.sr.ht/~tslocum/netris/pkg/mino"
)

func TestRenderMatrix(t *testing.T) {
	m, err := mino.NewTestMatrix()
	if err != nil {
		t.Error(err)
	}

	m.AddTestBlocks()

	var renderedMatrix []byte
	renderedMatrix = renderMatrix(m)

	_ = renderedMatrix
}

func BenchmarkRenderStandardMatrix(b *testing.B) {
	m, err := mino.NewTestMatrix()
	if err != nil {
		b.Error(err)
	}

	m.AddTestBlocks()

	var renderedMatrix []byte
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		renderedMatrix = renderMatrix(m)
	}

	_ = renderedMatrix
}

func BenchmarkRenderLargeMatrix(b *testing.B) {
	blockSize = 2

	m, err := mino.NewTestMatrix()
	if err != nil {
		b.Error(err)
	}

	m.AddTestBlocks()

	var renderedMatrix []byte
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		renderedMatrix = renderMatrix(m)
	}

	_ = renderedMatrix

	blockSize = 1
}
