package capture

import (
	"bytes"
	"crypto/rand"
	"flag"
	"io"
	"testing"

	"appscope.net/mem"
)

type TestSegmentInfo struct {
	length, limit int
}

func fill(data []byte, offset, length int, val byte) {
	for i := 0; i < length; i++ {
		data[offset+i] = val
	}
}

// we test writing to Segments
func TestCache(t *testing.T) {
	flag.Parse()

	bufSizes := []float32{.001, 0.5, 1, 2}
	relativeLimits := []float32{0.5, 1, 2} // compared to buffer

	// define how much bytes do we need to handle that ?
	needBytes := 0
	for _, k := range bufSizes {
		for _, l := range relativeLimits {
			needBytes += int(mem.CBufSize * k * l)
		}
	}
	t.Logf("needBytes = %d", needBytes)
	// make random buffer
	originalBytes := make([]byte, needBytes)
	if _, err := rand.Read(originalBytes); err != nil {
		t.Fatal(err)
	}

	// fill(originalBytes, 0, 16, 0)
	// fill(originalBytes, 16, 16, 1)
	// fill(originalBytes, 32, 16, 2)

	// t.Log(originalBytes)

	cache := NewSegmentedStream(bytes.NewReader(originalBytes))

	// STEP 1 - fill the cache
	Segments := []TestSegmentInfo{}
	bufferedReader := cache.BufferedReader()

	for _, k := range bufSizes {
		SegmentBytes := int(k * mem.CBufSize)
		buf := make([]byte, SegmentBytes, SegmentBytes)

		for _, limit := range relativeLimits {
			limitBytes := int(k * mem.CBufSize * limit)
			cache.StartSegment(int64(limitBytes))
			Segments = append(Segments, TestSegmentInfo{length: SegmentBytes, limit: limitBytes})
		read_loop:
			for bytesRead := 0; bytesRead < SegmentBytes; {
				if n, err := bufferedReader.Read(buf[bytesRead:]); err != nil && err != io.EOF {
					_ = "breakpoint"
					t.Fatal(err)
				} else if err == io.EOF {
					break read_loop
				} else {
					bytesRead += n
				}
			}
			// t.Logf("For segment size=%d limit=%d lefover=%d read=%v", SegmentBytes, limitBytes, bufferedReader.Buffered(), buf)
		}
	}
	cache.Close()

	// STEP 2 - compare
	t.Log(Segments)
	offset := 0

	for SegmentId, Segment := range Segments {
		cacheDump := bytes.Buffer{}
		n, _, err := cache.FlushSegment(SegmentId, &cacheDump)

		expectedBytes := Segment.length
		if Segment.length > Segment.limit {
			expectedBytes = Segment.limit
		}
		if n != expectedBytes {
			t.Errorf("ERR Segment %d size=%d got=%d", SegmentId, expectedBytes, n)
		} else {
			t.Logf("OK  Segment %d size=%d got=%d", SegmentId, expectedBytes, n)
		}
		if err != nil {
			t.Error(err)
		}
		dump := cacheDump.Bytes()
		if bytes.Equal(originalBytes[offset:offset+expectedBytes], dump) == false {
			_ = "breakpoint"
			t.Errorf("Segment %d bytes (size=%d) don't match to [%d:%d]", SegmentId, len(dump), offset, offset+expectedBytes)
		}
		// t.Logf(">>>>>>> WAS=%v", originalBytes[offset:offset+expectedBytes])
		// t.Logf(">>>>>>> GOT=%v", dump)
		offset += Segment.length
	}

	// STEP 3 - avoid memory leaks
	cache.DiscardAll()
	if mem.GetAllocatedBuffersCount() != 0 {
		t.Errorf("After discarding cache, still have %d buffers allocated", mem.GetAllocatedBuffersCount())
	}
}
