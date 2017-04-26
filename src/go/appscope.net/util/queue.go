package util

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/golang/glog"
)

type Queue interface {
	ISampler
	Push(v interface{})
	Pop() interface{}
	Size() int64
}

const (
	cQueueSize             = int64(1024)
	cQueueSizeThreshold75  = (2 * cQueueSize) / 3
	cQueueSizeThreshold50  = cQueueSize / 2
	cQueueSamplingInterval = time.Second * 10
	cSelectorTimeout       = time.Second
)

type queue struct {
	name  string
	ch    chan interface{}
	count int64
}

func NewMonitoredQueue(name string) Queue {
	q := &queue{
		name: name,
		ch:   make(chan interface{}, cQueueSize),
	}
	AddGauge(q, cQueueSamplingInterval)
	return q
}

func (q *queue) Push(v interface{}) {
	q.ch <- v
	n := atomic.AddInt64(&q.count, 1)
	if n > cQueueSizeThreshold75 {
		glog.Errorf("Queue %s is over 75 percent : %d", q.name, n)
	} else if n > cQueueSizeThreshold50 {
		glog.Warningf("Queue %s is over 50 percent : %d", q.name, n)
	}
}

func (q *queue) Size() int64 {
	return atomic.LoadInt64(&q.count)
}

func (q *queue) Pop() (v interface{}) {
	v = <-q.ch
	atomic.AddInt64(&q.count, -1)
	return
}

func (q *queue) Name() string {
	return q.name
}

func (q *queue) Sample() float64 {
	return float64(atomic.LoadInt64(&q.count))
}

/*
 not thread safe
*/
type Selector interface {
	// either returns result from one of the queues, or nil if timeout
	LoopWithTimeout(timeout time.Duration, onTimeout Handler)
	Loop()
}

type Handler func(val interface{}) error

type selector struct {
	q          []*queue
	h          []Handler
	selectCase []reflect.SelectCase
}

func NewSelector(qm map[Queue]Handler) Selector {
	s := &selector{
		q:          make([]*queue, 0, len(qm)+1),
		h:          make([]Handler, 0, len(qm)+1),
		selectCase: make([]reflect.SelectCase, 0, len(qm)+1),
	}

	for q, h := range qm {
		if qq, ok := q.(*queue); !ok {
			panic("Unsupported queue type")
		} else {
			s.q = append(s.q, qq)
			s.h = append(s.h, h)
			s.selectCase = append(s.selectCase, reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(qq.ch),
			})
		}
	}

	return s
}

func (s *selector) LoopWithTimeout(t time.Duration, onTimeout Handler) {
	if onTimeout != nil {
		s.selectCase = append(s.selectCase, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(time.Tick(t))})
		s.q = append(s.q, nil)
		s.h = append(s.h, onTimeout)
	}

	for { // TODO : handle channel closure
		i, recv, _ := reflect.Select(s.selectCase)
		if s.q[i] != nil {
			atomic.AddInt64(&s.q[i].count, -1)
		}
		if s.h[i] != nil {
			vv := recv.Interface()
			if e := SafeRun(func() {
				if e := s.h[i](vv); e != nil {
					glog.Errorf("%v(%v) returned %v", s.h[i], vv, e)
				}
			}); e != nil {
				glog.Errorf("%v(%v) panicked", s.h[i], vv)
			}
		}
	}
}

func (s *selector) Loop() {
	s.LoopWithTimeout(0, nil)
}

/*
 * balances the load between multiple allocated workers
**/
type JobDirector interface {
	Add(val interface{}, handler Handler, errch chan error)
}

func NewJobDirector(name string, nworkers int64) JobDirector {
	p := &pusher{
		q: make([]Queue, 0, nworkers),
	}

	selectorMap := map[Queue]Handler{}
	for i := int64(1); i <= nworkers; i++ {
		q := NewMonitoredQueue(fmt.Sprintf("%s-%d", name, i))
		p.q = append(p.q, q)
		selectorMap[q] = jobDirectorExecutor
	}

	p.selector = NewSelector(selectorMap)
	go SafeRun(p.selector.Loop)

	return p
}

type pusher struct {
	q        []Queue
	selector Selector
}

type jobItem struct {
	val     interface{}
	handler Handler
	errch   chan error
}

func jobDirectorExecutor(v interface{}) (err error) {
	item := v.(*jobItem)
	defer func() {
		if x := recover(); x != nil {
			p := RenderPanic(x)
			glog.Errorf("Panic %v(%v): %v", item.handler, item.val, p)
			item.errch <- ePanic
			err = ePanic
		}
	}()

	err = item.handler(item.val)
	item.errch <- err
	return err
}

func (p *pusher) Add(val interface{}, handler Handler, errch chan error) {
	minSize := int64(0)
	minIdx := 0

	for i, q := range p.q {
		s := q.Size()
		if s < minSize {
			minSize = s
			minIdx = i
		}
	}

	p.q[minIdx].Push(&jobItem{
		val:     val,
		handler: handler,
		errch:   errch,
	})
}
