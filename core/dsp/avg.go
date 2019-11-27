package dsp

import "math"

func newAverager(length, blockSize int) *averager {
	result := &averager{
		length:  length,
		buffer:  make([][]float64, length),
		index:   0,
		current: make([]float64, blockSize),
	}
	for i := range result.buffer {
		result.buffer[i] = make([]float64, blockSize)
	}
	return result
}

type averager struct {
	length  int
	buffer  [][]float64
	index   int
	current []float64
}

func (a *averager) Put(row []float64) []float64 {
	for i := range row {
		a.current[i] += ((row[i] - a.buffer[a.index][i]) / float64(a.length))
	}
	a.buffer[a.index] = row
	a.index = (a.index + 1) % a.length
	return a.current
}

func newMaxer(length, blockSize int) *maxer {
	result := &maxer{
		length:  length,
		buffer:  make([][]float64, length),
		index:   0,
		current: make([]float64, blockSize),
	}
	for i := range result.buffer {
		result.buffer[i] = make([]float64, blockSize)
	}
	return result
}

type maxer struct {
	length  int
	buffer  [][]float64
	index   int
	current []float64
}

func (m *maxer) Put(row []float64) []float64 {
	for i := range row {
		m.current[i] = math.Max(m.buffer[m.index][i], row[i])
	}
	m.buffer[m.index] = row
	m.index = (m.index + 1) % m.length
	return m.current
}

func newSlidingWindow(length int) *slidingWindow {
	result := &slidingWindow{
		length:  length,
		buffer:  make([]float64, length),
		index:   0,
		current: 0,
	}
	return result
}

type slidingWindow struct {
	length  int
	buffer  []float64
	index   int
	current float64
}

func (w *slidingWindow) Put(v float64) float64 {
	w.current += ((v - w.buffer[w.index]) / float64(w.length))
	w.buffer[w.index] = v
	w.index = (w.index + 1) % w.length
	return w.current
}

func newSlidingMax(length int) *slidingMax {
	result := &slidingMax{
		length:      length,
		buffer:      make([]float64, length),
		bufferIndex: 0,
		maxIndex:    0,
		index:       0,
	}
	return result
}

type slidingMax struct {
	length      int
	buffer      []float64
	bufferIndex int
	maxIndex    int
	index       int
}

func (m *slidingMax) Put(v float64) int {
	currentMaxInBufferIndex := ((m.length + m.bufferIndex) - (m.index - m.maxIndex)) % m.length
	if v >= m.buffer[currentMaxInBufferIndex] {
		m.maxIndex = m.index
	}
	m.buffer[m.bufferIndex] = v

	m.bufferIndex = (m.bufferIndex + 1) % m.length
	m.index++
	return m.maxIndex
}
