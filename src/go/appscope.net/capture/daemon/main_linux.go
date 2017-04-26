package main

/*
#cgo LDFLAGS: -lnuma
*/

import (
	"appscope.net/capture"
	"appscope.net/ews"
	"appscope.net/filter"
	"appscope.net/mem"
	"appscope.net/tlsproxy"
	"appscope.net/util"

	"flag"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/golang/glog"

	"github.com/google/gopacket"

	_ "net/http/pprof"
)

var iface              = flag.String("i",           "eth0", "Interface to get packets from")
var enableKernelBpf    = flag.Bool("kernelbpf",     false,  "must be false on any virtualized environment")
var enableTcpFiltering = flag.Bool("tcpfilter",     false,  "Control TCP HTTP filtering")
var enableTlsFiltering = flag.Bool("tlsfilter",     false,  "Control TLS HTTP filtering")
var certsPath          = flag.String("certsPath",   "",     "Certifacates storage path")
var excludedDomains    = flag.String("exclDomains", "",     "Comma-separated list of excluded TLS domains")
var version string

func main() {        
	flag.Parse()
	
	glog.Infof("App version %s", version)

	go func() {
		glog.Fatal(http.ListenAndServe(":6060", nil))
	}()

	ip, err := getInterfaceIP(*iface)
	if err != nil {
		glog.Fatal("Cannot get IP address of interface %v : %v", *iface, err)
	}

	if true {
		sample_interval := time.Duration(5 * time.Second)
		util.AddGauge(mem.HeapSysSampler, sample_interval)
		util.AddGauge(mem.HeapAllocSampler, sample_interval)
		util.AddGauge(mem.HeapIdleSampler, sample_interval)
		util.AddGauge(mem.HeapReleasedSampler, sample_interval)
	}

	tracker := capture.NewConnectionTracker(map[gopacket.Endpoint]bool{util.HUint16toPort(80): true}, ip)
	capture.RegisterEws(tracker)

	filterController := filter.NewController()
	filter.RegisterEws(filterController)

	ews.Run(":8866")
	
	/*
		glog.Infof("Starting PF_RING capture on %v IP=%v", *iface, ip)
		go util.SafeRun(func() { capture.RunPFRing(tracker, *iface, ip, *enableKernelBpf) })
	*/

	tlsPxy := tlsproxy.TLSProxy{
		LocalAddr         : flag.String("tls", ":8443", "TLS proxy port"),
		Tracker           : tracker,
		FilterController  : filterController,
		EnableFiltering   : *enableTlsFiltering,
		StaticExclByDomain: getExcludedDomains(*excludedDomains),
	}

	glog.Infof("Starting TLS Proxy on %v", tlsPxy.LocalAddr)
	tlsproxy.RegisterEws(&tlsPxy)
	
	go util.SafeRun(func() {
		if err := tlsPxy.Listen(*certsPath); err != nil {
			glog.Fatalf("TLS Proxy could not start : %v", err)
		}
	})

	tcpPxy := tlsproxy.TCPProxy{
		LocalAddr       : flag.String("tcp", ":8080", "TCP proxy port"),
		Tracker         : tracker,
		FilterController: filterController,
		EnableFiltering : *enableTcpFiltering,
	}

	glog.Infof("Starting TCP Proxy on %v", tcpPxy.LocalAddr)
	go util.SafeRun(func() {
		if err := tcpPxy.Listen(); err != nil {
			glog.Fatalf("TCP Proxy could not start : %v", err)
		}
	})

	<-make(chan bool)

}

func getInterfaceIP(eth string) (net.IP, error) {
	if ifce, err := net.InterfaceByName(eth); err != nil {
		return nil, err
	} else if addrs, err := ifce.Addrs(); err != nil {
		return nil, err
	} else {
		for _, addr := range addrs {
			if ip, ok := (addr).(*net.IPNet); ok { // if it's nill, then it's i.e. MAC
				if ipv4 := ip.IP.To4(); ipv4 != nil {
					return ipv4, nil
				}
			} else {
				glog.Info(ip, addr)
			}
		}
	}
	return nil, fmt.Errorf("No IPv4 addresses assigned to %v", eth)
}

func getExcludedDomains(domainsArg string) (excludedDomains map[string] bool) {
	excludedDomains = make(map[string] bool)
	domains := strings.Split(domainsArg, ",")
	
	for i := 0; i < len(domains); i++ {
		excludedDomains[domains[i]] = true
		glog.Infof("Excluding domain: %v", domains[i])
	}
	
	return excludedDomains
}
