package mem

/*
 transparently mirrors everything it reads and performs statistics calculation
*/

import (
	"fmt"
	"io"
)

const (
	cChunks = 10
)

type ProxyReader struct {
	reader         io.Reader
	onComplete     func(bytes int)
	totalBytesRead int

	chunks   [][]byte
	maxBytes int // maximum to keep in this cache
	nBytes   int // currently bytes cached
}

func MakeProxyReader(r io.Reader,
	maxBytes int, hook func(bytes int)) *ProxyReader {
	return &ProxyReader{
		reader:     r,
		onComplete: hook,

		chunks:   make([][]byte, 0, cChunks),
		maxBytes: maxBytes,
	}
}

func (r *ProxyReader) BytesCached() int {
	return r.nBytes
}

func (r *ProxyReader) BytesProxied() int {
	return r.totalBytesRead
}

func (r *ProxyReader) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("{\"total\":%d, \"saved\":%d}", r.totalBytesRead, r.nBytes)), nil
}

func (r *ProxyReader) Read(b []byte) (n int, err error) {
	n, err = r.reader.Read(b)
	if n > 0 {
		r.totalBytesRead += n
		r.append(b[:n])
	}
	return n, err
}

func (r *ProxyReader) Close() error {

	if r.onComplete != nil {
		r.onComplete(r.totalBytesRead)
	}

	if closer, ok := r.reader.(io.Closer); ok {
		return closer.Close()
	} else {
		return nil
	}
}

func (r *ProxyReader) Purge() {
	for i := 0; i < len(r.chunks); i++ {
		RecycleBuffer(r.chunks[i])
	}
	r.nBytes = 0
}

func (r *ProxyReader) Flush(out io.Writer) (bytesDumped int, err error) {
	n := 0
	bytesDumped = 0
	for i := 0; i < len(r.chunks); i++ {
		if i == len(r.chunks)-1 {
			n, err = out.Write(r.chunks[i][:(r.nBytes - i*CBufSize)])
		} else {
			n, err = out.Write(r.chunks[i])
		}
		bytesDumped += n
		RecycleBuffer(r.chunks[i])
		r.chunks[i] = nil
	}
	return
}

func (r *ProxyReader) append(buf []byte) {
	for p := 0; p < len(buf) && p < r.maxBytes && r.nBytes < r.maxBytes; {
		idx := r.nBytes / CBufSize
		i := r.nBytes % CBufSize
		if len(r.chunks) == idx {
			r.chunks = append(r.chunks, GetBuffer())
		}
		n := copy(r.chunks[idx][i:], buf[p:])
		r.nBytes += n
		p += n

		if r.nBytes > r.maxBytes {
			r.nBytes = r.maxBytes
		}
	}
}
