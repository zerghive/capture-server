package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"

	"appscope.net/tlsproxy"

	"github.com/golang/glog"
)

var certsPath = flag.String("certsPath", ".", "Certifacates storage path.")

func main() {
	go func() {
		glog.Fatal(http.ListenAndServe(":6061", nil))
	}()

	flag.Parse()
	glog.Infof("Starting up!")

	pxy := tlsproxy.TLSProxy{
		LocalAddr: flag.String("l", ":8443", "local address"),
	}

	if err = pxy.Listen(*certsPath); err != nil {
		glog.Fatal(err)
	}
}
