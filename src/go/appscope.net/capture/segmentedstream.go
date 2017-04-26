// +build ignore
package capture

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"sync"
	"time"

        "appscope.net/filter"
	"appscope.net/mem"
	"github.com/golang/glog"
)

const (
	CSegmentedStreamNoLimit      = math.MaxInt32
	CSegmentedStreamDefaultLimit = 2 << 19
	cChunks                      = 12
)

type segmentInfo struct {
	offset, nBytes, maxBytes, processedBytes int64 // offset is against real
	start                                    *time.Time
	duration                                 time.Duration

	payload interface{}

	parseError error
}

type SegmentedStream struct {
	closed       bool
	reader       io.Reader
	writer       io.Writer
	closeOnClose bool

	chunks              [][]byte
	nBytes              int64 // defines how many bytes currently the channel holds
	totalBytesProcessed int64 // total amount of bytes proxies (but not necessarily cached)

	segmentsMutex       sync.RWMutex
	segments            []*segmentInfo
	bufferedReader      *bufio.Reader

	overlapBuffer                           []byte
	overlapBufferBytes, overlapBufferOffset int
	
	modifier *filter.Modifier
}

/*

 to avoid extensive use of locking unlocking, the following design decision is taken:

  - segments implicitly own byte chunks
  - segments are only modified by:
    + start Segment (called from assembler)
    + discard Segment (called from exporter & trimmer processes)
  - byte chunks are released whenever its done

  closeOnClose has special meaning, depending whether we'll be indexing full duplex connections or not
  i.e. in case PFRING, we do need close reader on exit. However we shouldn't in TLS socket proxying case
*/

func NewSegmentedStream(in io.Reader, out io.Writer, closeOnClose bool) *SegmentedStream {
	ch := &SegmentedStream{
		closed:       false,
		reader:       in,
		writer:       out,
		closeOnClose: closeOnClose,
		chunks:       make([][]byte, 0, cChunks),
		nBytes:       0,
		segments:     make([]*segmentInfo, 0, cChunks),
	}

	return ch
}

func (ch *SegmentedStream) BufferedReader() *bufio.Reader {
	if ch.bufferedReader == nil {
		ch.bufferedReader = bufio.NewReader(ch)
	}

	return ch.bufferedReader
}

/*
 marks beginning of a Segment
 a correction is applied due to the use of buffered stream
*/
func (ch *SegmentedStream) StartSegment(limit int64) (id int) {
	ch.segmentsMutex.Lock()
	defer ch.segmentsMutex.Unlock()
	
	ts := time.Now()
	timestamp := &ts

	if glog.V(vPacketTrace) {
		glog.Infof("New segment id=%d, limit=%d", len(ch.segments), limit)
	}
	buffered := int64(ch.BufferedReader().Buffered())

	if buffered > 0 { // exhaust buffer and subtract what's wrong
		prevIndex := len(ch.segments) - 1

		if glog.V(vPacketTrace) {
			glog.Infof(" - BeforeCorrection, buffered=%d, currentSegment(id=%d,nB=%d,pB=%d) ch(nB=%d,pB=%d)",
				buffered, prevIndex,
				ch.segments[prevIndex].nBytes, ch.segments[prevIndex].processedBytes,
				ch.nBytes, ch.totalBytesProcessed)
		}

		if ch.segments[prevIndex].processedBytes < buffered {
			panic("buffer larger then chunk")
		}

		ch.totalBytesProcessed -= buffered
		ch.segments[prevIndex].processedBytes -= buffered

		if ch.segments[prevIndex].processedBytes < ch.segments[prevIndex].nBytes {
			ch.nBytes -= ch.segments[prevIndex].nBytes - ch.segments[prevIndex].processedBytes
			ch.segments[prevIndex].nBytes = ch.segments[prevIndex].processedBytes
		}

		if glog.V(vPacketTrace) {
			glog.Infof(" - AfterCorrection, currentSegment(id=%d,nB=%d,pB=%d) ch(nB=%d,pB=%d)",
				prevIndex, ch.segments[prevIndex].nBytes, ch.segments[prevIndex].processedBytes,
				ch.nBytes, ch.totalBytesProcessed)
		}

		if ch.overlapBuffer != nil {
			if glog.V(vPacketTrace) {
				glog.Infof(" - Overlap still here [%d:%d]", ch.overlapBufferOffset, ch.overlapBufferBytes)
			}
		} else {
			ch.overlapBuffer = mem.GetBuffer()
			if n, err := ch.bufferedReader.Read(ch.overlapBuffer[:buffered]); int64(n) != buffered || err != nil {
				panic(fmt.Sprintf("toAppend=%d, n=%d, err=%v", buffered, n, err))
			} else {
				ch.overlapBufferBytes = n
				ch.overlapBufferOffset = 0
			}
		}
	} else {
		timestamp = nil
	}

	ch.segments = append(ch.segments, &segmentInfo{
		offset:         ch.nBytes,
		nBytes:         0,
		processedBytes: 0,
		maxBytes:       limit,
		start:          timestamp,
	})

	return len(ch.segments) - 1
}

func (ch *SegmentedStream) GetSegmentTime(id int) (time.Time, time.Duration) {
	if id >= len(ch.segments) {
		glog.Errorf("GET segment[%d] failed", ch, id)
		return time.Time{}, 0
	}
	
	segment := ch.segments[id]
	
	if segment.start == nil {
		glog.Errorf("GET segment[%d] time failed", ch, id)
		return time.Time{}, 0
	}
	
	return *segment.start, segment.duration
}

func (ch *SegmentedStream) SetPayload(payload interface{}) {
	id := len(ch.segments) - 1
	ch.segments[id].payload = payload

	// log.Printf("%p SET segment[%d].payload=%v", ch, id, payload)
}

var eNoSegment = fmt.Errorf("no such segment")

func (ch *SegmentedStream) GetPayload(id int) (interface{}, error) {
	if id >= len(ch.segments) {
		glog.Errorf("GET segment[%d] failed", ch, id)
		return nil, eNoSegment
	}

	segment := ch.segments[id]
	if segment.parseError != nil && glog.V(vPacketTrace) {
		glog.Infof("GET segment[%d].payload : has error=%v, nBytes=%d", id, segment.parseError, segment.nBytes)
	}
		
	return segment.payload, segment.parseError
}

// marks current segment as unparsed; this stream is broken now
func (ch *SegmentedStream) ParseFailed(err error) {
	ch.segments[len(ch.segments)-1].parseError = err
}

func (ch *SegmentedStream) IsFailed() bool {
	return ch.segments[len(ch.segments)-1].parseError != nil
}

func (ch *SegmentedStream) MarkDuration() (time.Time, time.Duration) {
	id := len(ch.segments) - 1
	if ch.segments[id].start == nil {
		ts := time.Now()
		ch.segments[id].start = &ts
	}

	ch.segments[id].duration = time.Since(*ch.segments[id].start)

	return *ch.segments[id].start, ch.segments[id].duration
}

func (ch *SegmentedStream) GetEvenSegments(from, to time.Time) (segmentIDs []int) {
	segmentIDs = make([]int, 0, len(ch.segments))
	for id, segment := range ch.segments {
		if segment != nil && segment.start != nil {
			if (segment.start.Before(to) && segment.start.After(from)) && (id%2 == 0) && (segment.parseError == nil) {
				segmentIDs = append(segmentIDs, id)
			}
		}
	}
	return
}

func (ch *SegmentedStream) GetTimeRange() (*time.Time, *time.Time) {
	last := len(ch.segments) - 1
	if last < 0 {
		return nil, nil
	}

	return ch.segments[0].start, ch.segments[last].start
}

func (ch *SegmentedStream) Destroy() {
	if !ch.closed {
		panic("Cannot recycle a stream which is still open")
	}
	for i, _ := range ch.chunks {
		mem.RecycleBuffer(ch.chunks[i])
		ch.chunks[i] = nil
	}
	ch.segments = nil
}

// writes data within a given Segment
func (ch *SegmentedStream) FlushSegment(id int, out io.Writer) (bytesDumped, bytesProcessed int, err error) {
	ch.segmentsMutex.Lock()
	defer ch.segmentsMutex.Unlock()
	
	if id >= len(ch.segments) {
		glog.Errorf("Flush segment %d failed", id)
		return 0, 0, eNoSegment
	}

	segment := ch.segments[id]

	bytesProcessed = int(segment.processedBytes)

	if segment == nil {
		return 0, 0, fmt.Errorf("Cannot save segment[%d]=<nil>", id)
	}

	n := 0
	bytesDumped = 0
	startIndex := segment.offset / mem.CBufSize
	endIndex := (segment.offset + segment.nBytes) / mem.CBufSize

	for i := startIndex; i <= endIndex; i++ {
		chunkOffset := int64(0)
		copy_up_to := mem.CBufSize

		if i == startIndex {
			chunkOffset = segment.offset % mem.CBufSize
		}
		if i == endIndex {
			copy_up_to = int((segment.offset + segment.nBytes) % mem.CBufSize)
		}
		n, err = out.Write(ch.chunks[i][chunkOffset:copy_up_to])

		bytesDumped += n
	}

	return
}

func (ch *SegmentedStream) GetLatestDataPoint() {

}

// identify which chunks have been freed already and release them
func (ch *SegmentedStream) recycleChunks() {
	lastUsedChunk := -1

	segmentsPresent := false
	for _, segment := range ch.segments {
		if segment != nil { // release all chunks before this one
			segmentsPresent = true
			for i := lastUsedChunk + 1; i < int((segment.offset+1)/mem.CBufSize); i++ {
				if ch.chunks[i] != nil {
					mem.RecycleBuffer(ch.chunks[i])
					ch.chunks[i] = nil
				}
			}
			lastUsedChunk = int((segment.offset + segment.nBytes + 1) / mem.CBufSize)
		}
	}

	if !segmentsPresent {
		for i, _ := range ch.chunks {
			if ch.chunks[i] != nil {
				mem.RecycleBuffer(ch.chunks[i])
				ch.chunks[i] = nil
			}
		}
	}

}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func (ch *SegmentedStream) readCopy(data []byte) (int, *time.Time, error) {
	bytesRead, readErr := ch.reader.Read(data)

	timestamp := time.Now()

	if ch.writer == nil {
		return bytesRead, &timestamp, readErr
	}

	for i := 0; i < bytesRead; {
		if bytesWritten, writeErr := ch.writer.Write(data[i:bytesRead]); writeErr != nil {
			glog.Errorf("incomplete copy : %v", writeErr)
			ch.writer = nil // closing would be handled as part of overall Close() call
			return bytesRead, &timestamp, writeErr
		} else {
			i += bytesWritten
		}
	}

	return bytesRead, &timestamp, readErr
}

// io.ReadCloser implementation

func (ch *SegmentedStream) Read(b []byte) (n int, err error) {
	data := b[:]

	segmentId := len(ch.segments) - 1

	if ch.overlapBuffer != nil {
		n = min(len(b), (ch.overlapBufferBytes - ch.overlapBufferOffset))
		if glog.V(vPacketTrace) {
			glog.Infof("Read segment=%d in=%d, read %d from overlap=[%d:%d]", segmentId, len(b), n,
				ch.overlapBufferOffset, ch.overlapBufferBytes)
		}
		data = ch.overlapBuffer[ch.overlapBufferOffset:n]
		copy(b, data)
		ch.overlapBufferOffset += n
	} else {
		var timestamp *time.Time
		n, timestamp, err = ch.readCopy(data)

		if ch.segments[segmentId].start == nil {
			ch.segments[segmentId].start = timestamp
		}

		if glog.V(vPacketTrace) {
			glog.Infof("Read segment=%d in=%d, read %d from stream", segmentId, len(b), n)
		}
	}
	//log.Printf(" - data=%v", data[:n])
	if n > 0 {
		ch.totalBytesProcessed += int64(n)
		ch.append(data[:n])
	}

	if ch.overlapBufferOffset > ch.overlapBufferBytes { // never happens
		panic(fmt.Sprintf("invalid offset=%d length=%d", ch.overlapBufferOffset, ch.overlapBufferBytes))
	} else if ch.overlapBufferOffset == ch.overlapBufferBytes && ch.overlapBuffer != nil {
		mem.RecycleBuffer(ch.overlapBuffer)
		ch.overlapBuffer = nil
		ch.overlapBufferOffset = 0
		ch.overlapBufferBytes = 0
	}

	return n, err
}

func (ch *SegmentedStream) Close() error {
	ch.closed = true

	ch.StartSegment(0)

	if ch.overlapBuffer != nil {
		mem.RecycleBuffer(ch.overlapBuffer)
		ch.overlapBuffer = nil
		ch.overlapBufferBytes = 0
		ch.overlapBufferOffset = 0
	}

	if ch.closeOnClose {
		if closer, ok := ch.reader.(io.Closer); ok {
			eReader := closer.Close()
			if eReader != nil && eReader != io.EOF {
				glog.Warning(eReader)
			}
		}
		if closer, ok := ch.writer.(io.Closer); ok && closer != nil {
			eWriter := closer.Close()
			if eWriter != nil && eWriter != io.EOF {
				glog.Warning(eWriter)
			}
		}
	}

	return nil
}

/*
 appends data to the given chunk but not exceeding maximum if set
*/
func (ch *SegmentedStream) append(buf []byte) {
	segment_id := len(ch.segments) - 1
	//log.Printf("append id=%d data=%v", segment_id, buf[:8])

	segment := ch.segments[segment_id]
	segment.processedBytes += int64(len(buf))

	for p := 0; p < len(buf) && segment.nBytes < segment.maxBytes; {
		idx := int((segment.offset + segment.nBytes) / mem.CBufSize)
		i := (segment.offset + segment.nBytes) % mem.CBufSize
		if len(ch.chunks) == idx {
			ch.chunks = append(ch.chunks, mem.GetBuffer())
		}
		// check we don't exceed maxBytes
		copy_up_to := len(buf)
		if (int64(p+copy_up_to) + segment.nBytes) > segment.maxBytes {
			copy_up_to = min(int(segment.maxBytes-segment.nBytes)+p, copy_up_to)
		}
		if glog.V(vPacketTrace) {
			glog.Infof("Store(id=%d,max=%d,given=%d) chunks[%d][%d:] from buf[%d:%d]", segment_id, segment.maxBytes, len(buf),
				idx, i, p, copy_up_to)
		}

		n := copy(ch.chunks[idx][i:], buf[p:copy_up_to])

		ch.nBytes += int64(n)
		segment.nBytes += int64(n)
		p += n
	}

}

func (ch *SegmentedStream) RegisterModifier(modifier *filter.Modifier) {
    ch.modifier = modifier
}

