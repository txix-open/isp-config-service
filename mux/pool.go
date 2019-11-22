package mux

import (
	"sync"
)

// could be increased if needed for new matchers
const bufferSize = 10

var pool sync.Pool

func getBuffer() []byte {
	b, ok := pool.Get().([]byte)
	if !ok {
		b = make([]byte, bufferSize)
	}
	return b
}

func putBuffer(b []byte) {
	pool.Put(b[:cap(b)])
}
