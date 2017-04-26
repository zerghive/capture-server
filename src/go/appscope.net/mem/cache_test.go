package mem

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"testing"
)

func TestCache(t *testing.T) {
	if plot, err := RunRecyclerProfilePlot(); err != nil {
		t.Error(err)
	} else {
		t.Log("Recycler Plot:", plot.GetUrl())
	}

	// no buffers should be out by now
	if buffersOut != 0 {
		t.Errorf("buffersOut = %d", buffersOut)
	}

	kk := []float32{0, 1.0 / cBufSize, 0.5, 1.0, 1.5, 1.7, 2.0, 2.3, 2.5, 2.7, 3}
	for _, k1 := range kk {
		for _, k2 := range kk {
			norig := int(cBufSize * k1)
			nmax := int(cBufSize * k2)
			testCache(t, norig, nmax)
		}
	}

	// no buffers should be out by now
	if buffersOut != 0 {
		t.Errorf("buffersOut = %d", buffersOut)
	}
}

func testCache(t *testing.T, nOriginalBytes, nMaxBytes int) {
	TAG := fmt.Sprintf("testCache(nOriginalBytes=%d, nMaxBytes=%d) ", nOriginalBytes, nMaxBytes)

	if buffersOut != 0 {
		t.Errorf("%s buffersOut = %d", TAG, buffersOut)
	}

	originalBytes := make([]byte, nOriginalBytes)
	if _, err := rand.Read(originalBytes); err != nil {
		t.Fatal(err)
	}

	original := bytes.NewReader(originalBytes)

	var hookReportedBytes int
	cache := MakeProxyReader(original, nMaxBytes, func(n int) {
		hookReportedBytes = n
	})

	out := bytes.Buffer{}
	if nCopied, err := io.Copy(&out, cache); err != nil {
		t.Fatal(TAG, err)
	} else if nCopied != int64(nOriginalBytes) {
		t.Errorf("%s nCopied(%d) != nOriginalBytes", TAG, nCopied, nOriginalBytes)
	}

	// compare data copied
	if bytes.Equal(originalBytes, out.Bytes()) == false {
		t.Error(TAG, "originalBytes != out.Bytes()")
	}

	// check hook on close reporting
	cache.Close()
	if hookReportedBytes != nOriginalBytes {
		t.Errorf("%s hookReportedBytes(%d) != nOriginalBytes(%d)", TAG, hookReportedBytes, nOriginalBytes)
	}

	// now flush cache and compare it's correct
	cacheDump := bytes.Buffer{}
	if nFlushed, err := cache.Flush(&cacheDump); err != nil {
		t.Error(TAG, err)
	} else if nMaxBytes >= nOriginalBytes && nFlushed != nOriginalBytes {
		t.Errorf("%s nFlushed(%d) != nOriginalBytes(%d)", TAG, nFlushed, nOriginalBytes)
	} else if nMaxBytes < nOriginalBytes && nFlushed != nMaxBytes {
		t.Errorf("%s nFlushed(%d) != nMaxBytes(%d) len=%d cache.nBytes=%d", TAG, nFlushed, nMaxBytes, len(cache.chunks), cache.nBytes)
	}

	// finally check contents match
	cacheDumpBytes := cacheDump.Bytes()
	if bytes.Equal(cacheDumpBytes, originalBytes[:len(cacheDumpBytes)]) == false {
		t.Error(TAG, "cacheDumpBytes don't match original")
	}

	// check we did not leak any buffers
	if buffersOut != 0 {
		t.Errorf("%s buffersOut = %d", TAG, buffersOut)
	}
}
