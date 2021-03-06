package util

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"

	"github.com/golang/glog"
)

var ePanic = errors.New("Panic in goroutine")

func SafeRun(f func()) (err error) {
	defer func() {
		if x := recover(); x != nil {
			p := RenderPanic(x)
			glog.ErrorDepth(1, p)
			err = ePanic
		}
	}()

	f()
	return nil
}

func RenderPanic(x interface{}) string {
	buf := make([]byte, 16<<10) // 16 KB should be plenty
	buf = buf[:runtime.Stack(buf, false)]

	if false {
		return fmt.Sprintf("%v : %v", x, string(buf))
	}

	// Remove the first few stack frames:
	//   this func
	//   the recover closure in the caller
	// That will root the stack trace at the site of the panic.
	const (
		skipStart  = "util.RenderPanic"
		skipFrames = 2
	)
	start := bytes.Index(buf, []byte(skipStart))
	p := start
	for i := 0; i < skipFrames*2 && p+1 < len(buf); i++ {
		p = bytes.IndexByte(buf[p+1:], '\n') + p + 1
		if p < 0 {
			break
		}
	}
	if p >= 0 {
		// buf[start:p+1] is the block to remove.
		// Copy buf[p+1:] over buf[start:] and shrink buf.
		copy(buf[start:], buf[p+1:])
		buf = buf[:len(buf)-(p+1-start)]
	}

	// Add panic heading.
	head := fmt.Sprintf("panic: %v\n\n", x)
	if len(head) > len(buf) {
		// Extremely unlikely to happen.
		return head
	}
	copy(buf[len(head):], buf)
	copy(buf, head)

	return string(buf)
}
