package mem

import (
	"appscope.net/util"
	"container/list"
	"sync/atomic"
	"time"
)

func init() {
	go util.SafeRun(recycleLoop)
}

const cDoRecycle = false

var makes uint64
var frees uint64
var buffersOut int32

const (
	CBufSize   = 8 << 10
	cGCTimeout = time.Minute
)

var get = make(chan ([]byte))
var give = make(chan ([]byte))

func makeBuffer() []byte {
	atomic.AddUint64(&makes, 1)
	return make([]byte, CBufSize)
}

type queued struct {
	when  time.Time
	slice []byte
}

func GetBuffer() []byte {
	atomic.AddInt32(&buffersOut, 1)
	if cDoRecycle {
		return <-get
	} else {
		return makeBuffer()
	}
}

func RecycleBuffer(buf []byte) {
	if buf == nil {
		panic("got nil buffer for recycle")
	}

	if cDoRecycle {
		give <- buf
	}
	atomic.AddInt32(&buffersOut, -1)
}

func GetAllocatedBuffersCount() int32 {
	return atomic.LoadInt32(&buffersOut)
}

func recycleLoop() {
	q := new(list.List)
	for {
		if q.Len() == 0 {
			q.PushFront(queued{when: time.Now(), slice: makeBuffer()})
		}

		e := q.Front()

		timeout := time.NewTimer(time.Minute)
		select {
		case b := <-give:
			timeout.Stop()
			q.PushFront(queued{when: time.Now(), slice: b})

		case get <- e.Value.(queued).slice:
			timeout.Stop()
			q.Remove(e)

		case <-timeout.C:
			e := q.Front()
			for e != nil {
				n := e.Next()
				if time.Since(e.Value.(queued).when) > time.Minute {
					q.Remove(e)
					e.Value = nil
					frees++
				}
				e = n
			}
		}
	}
}

func RunRecyclerProfilePlot() (*RuntimePlot, error) {
	return NewRuntimePlot("Recycler",
		[]util.ISampler{
			&FuncSampler{"Frees", func() float64 { return float64(frees) }},
			&FuncSampler{"Makes", func() float64 { return float64(makes) }},
			&FuncSampler{"Out there", func() float64 { return float64(buffersOut) }},
		},
		[]string{
			"26pqtycscz",
			"5pnscito40",
			"tbfrvzw91w"},
		time.Second*10)
}
