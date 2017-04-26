package mem

import (
	"reflect"
	"runtime"
)

var HeapAllocSampler = makeSampler("HeapAlloc")
var HeapIdleSampler = makeSampler("HeapIdle")
var HeapSysSampler = makeSampler("HeapSys")
var HeapReleasedSampler = makeSampler("HeapReleased")

type memSampler struct {
	name string
	stat runtime.MemStats
}

func makeSampler(field string) *memSampler {
	ms := memSampler{}
	ms.name = field

	if reflect.ValueOf(ms.stat).FieldByName(field).IsValid() == false {
		panic(field)
	}
	return &ms
}

func (mp *memSampler) Sample() float64 {
	runtime.ReadMemStats(&mp.stat)
	return float64(reflect.ValueOf(mp.stat).FieldByName(mp.name).Uint())
}

func (mp *memSampler) Name() string {
	return "Go_" + mp.name
}
