package mem

import (
	"math/rand"
	"net/http"
	"testing"
	"time"

	_ "net/http/pprof"
)

func TestRecycler(t *testing.T) {
	go func() {
		t.Fatal(http.ListenAndServe("localhost:6060", nil))
	}()

	stop := false
	go makejunk(&stop, t)

	if plot, err := RunMemProfilePlot(); err != nil {
		t.Error(err)
	} else {
		t.Log("RAM Plot:", plot.GetUrl())
	}

	<-time.After(time.Minute * 10)
	stop = true
	t.Logf("%d %d %d \n", makes, frees, getBufferMiss)
}

func makejunk(stop *bool, t *testing.T) {
	pool := make([][]byte, 10000)

	for !*stop {
		b := GetBuffer()
		i := rand.Intn(len(pool))
		if pool[i] != nil {
			RecycleBuffer(pool[i])
		}

		pool[i] = b

		time.Sleep(time.Millisecond)

	}
}
