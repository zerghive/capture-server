package mem

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"appscope.net/util"
)

const (
	cMaxDataPoints = 500
)

type FuncSampler struct {
	SampleName string
	SampleFn   func() float64
}

func (fs *FuncSampler) Sample() float64 {
	return reflect.ValueOf(fs.SampleFn).Call([]reflect.Value{})[0].Float()
}

func (fs *FuncSampler) Name() string {
	return fs.SampleName
}

type Trace struct {
	sampler util.ISampler
	token   string
	conn    net.Conn
	val     float64
}

func (mt *Trace) writeSample(xVal string) {
	str := fmt.Sprintf("{\"x\":\"%s\",\"y\":%f}\n", xVal, mt.val)
	if _, e := fmt.Fprintf(mt.conn, "%x\r\n%s\r\n", len(str), str); e != nil {
		log.Println("writeSample(", str, ")", e)
	}
}

func (mt *Trace) connect() error {
	conn, err := net.Dial("tcp", "stream.plot.ly:80")
	if err != nil {
		log.Print(err)
		return err
	}

	fmt.Fprintf(conn, "POST / HTTP/1.1\r\n")
	fmt.Fprintf(conn, "Host: stream.plot.ly\r\n")
	fmt.Fprintf(conn, "Transfer-Encoding: chunked\r\n")
	fmt.Fprintf(conn, "Connection: close\r\n")
	fmt.Fprintf(conn, "plotly-streamtoken: %s\r\n", mt.token)
	fmt.Fprintf(conn, "\r\n")

	mt.conn = conn

	return nil
}

type RuntimePlot struct {
	traces   []*Trace
	interval time.Duration
	url      string
}

func (plot *RuntimePlot) GetUrl() string {
	return plot.url
}

func NewRuntimePlot(PlotName string, Samplers []util.ISampler, Tokens []string, interval time.Duration) (*RuntimePlot, error) {
	if len(Samplers) != len(Tokens) {
		return nil, fmt.Errorf("Samplers[%d] don't match Tokens[%d] count", len(Samplers), len(Tokens))
	}

	plot := RuntimePlot{make([]*Trace, len(Samplers)), interval, ""}

	for i, _ := range plot.traces {
		plot.traces[i] = &Trace{Samplers[i], Tokens[i], nil, 0}
	}

	args := makeArgs(plot.traces)

	resp, err := http.PostForm("https://plot.ly/clientresp",
		url.Values{
			"un":       {"denis.s.mishin"},
			"key":      {"q36gqdhq66"},
			"platform": {"go"},
			"origin":   {"plot"},
			"args":     {args},
			"kwargs": {j(map[string]interface{}{
				"filename": PlotName,
				"fileopt":  "overwrite",
			})},
		})

	if err != nil {
		return nil, err
	}

	var respBody bytes.Buffer
	if _, err := io.Copy(&respBody, resp.Body); err != nil {
		return nil, fmt.Errorf("Failed to parse log body %v", err)
	}

	respJSON := map[string]string{}
	if err := json.Unmarshal(respBody.Bytes(), &respJSON); err != nil {
		return nil, fmt.Errorf("Failed to parse log body %s : %s", string(respBody.Bytes()), err)
	}

	if err, ok := respJSON["error"]; ok && err != "" {
		return nil, fmt.Errorf("Server returned %s", string(respBody.Bytes()))
	}

	plot.url = respJSON["url"]

	// --- open streams
	for i, _ := range plot.traces {
		if e := plot.traces[i].connect(); e != nil {
			log.Printf("Could not connect to plot.ly streaming : %v", err)
			return nil, err
		}
		log.Println("conn =", plot.traces[i].conn)
		// go util.SafeRun(func() { dumpout(plot.traces[i]) })
		go dumpout(plot.traces[i])
	}

	// ----
	go util.SafeRun(func() {
		for {
			now := timenow()
			for i, _ := range plot.traces {
				plot.traces[i].val = plot.traces[i].sampler.Sample()
			}
			for i, _ := range plot.traces {
				plot.traces[i].writeSample(now)
			}
			time.Sleep(interval)
		}
	})

	return &plot, nil
}

func j(val interface{}) string {
	if data, err := json.Marshal(val); err != nil {
		log.Println(val, err)
		return ""
	} else {
		return string(data)
	}
}

func timenow() string {
	hr, min, sec := time.Now().Clock()
	return fmt.Sprintf("%d:%d:%d", hr, min, sec)
}

func makeArgs(mems []*Trace) string {
	args := []interface{}{}

	for _, mt := range mems {
		args = append(args, map[string]interface{}{
			"x":    []int{},
			"y":    []int{},
			"type": "scatter",
			"mode": "line markers",
			"name": mt.sampler.Name(),
			"stream": map[string]interface{}{
				"token":     mt.token,
				"maxpoints": cMaxDataPoints,
			}},
		)
	}
	return j(args)
}

func dumpout(trace *Trace) {
	buf := bufio.NewReader(trace.conn)
	log.Printf("dumpout %+v", trace)
	for {
		if data, _, err := buf.ReadLine(); err != nil {
			log.Println(trace.sampler.Name(), err)
			trace.conn.Close()
			return
		} else {
			log.Println(string(data))
		}
	}
}
