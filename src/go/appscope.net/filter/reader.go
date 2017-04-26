package filter

import (
)

type ReadFunc func(b []byte) (n int, err error)

type Reader struct {
    readFunc ReadFunc
}

func NewReader(f ReadFunc) *Reader {
    reader := &Reader{
        readFunc: f,
    }

    return reader
}

func (reader *Reader) Read(b []byte) (n int, err error) {
    return reader.readFunc(b)
}
