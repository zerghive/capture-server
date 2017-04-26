package capture

// #cgo LDFLAGS: -lnuma

import (
	"fmt"
	"net"
	"runtime"
	"time"

	"appscope.net/util"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pfring"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"

	"github.com/golang/glog"
)

const (
	PF_RING_USERSPACE_BPF = (1 << 13)
)

// takes interface name - i.e. "eth0"
// runs indefinitely
func RunPFRing(tracker *ConnectionTracker, eth string, ip net.IP, enableKernelBPF bool) error {
	bpf := fmt.Sprintf("tcp and host %s and port 80", ip)
	glog.Infof("BPF =%v", bpf)

	pfFlag := pfring.FlagPromisc
	if enableKernelBPF == false {
		glog.Infof("Kernel BPF is disabled")
		pfFlag |= PF_RING_USERSPACE_BPF
	}

	ring, err := pfring.NewRing(eth, 14000, pfFlag)
	if err != nil {
		return err
	} else if err = ring.SetSocketMode(pfring.ReadOnly); err != nil {
		return fmt.Errorf("pfring SetSocketMode error:", err)
	} else if err := ring.SetBPFFilter(bpf); err != nil {
		return fmt.Errorf("Set BPF %v", err)
	} else if err = ring.Enable(); err != nil {
		return fmt.Errorf("pfring Enable error: %v", err)
	}

	// Set up assembly
	streamFactory := NewHttpStreamFactory(tracker)
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)

	packetSource := gopacket.NewPacketSource(ring, layers.LinkTypeEthernet)
	packetSource.NoCopy = true
	packets := packetSource.Packets()
	ticker := time.Tick(time.Minute)

	runtime.LockOSThread()
	for {
		select {
		case packet := <-packets:
			if packet == nil {
				return nil
			}
			if packet.NetworkLayer() == nil || packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
				glog.Errorf("Unusable packet %v", packet)
				continue
			}
			tcp := packet.TransportLayer().(*layers.TCP)
			assembler.AssembleWithTimestamp(packet.NetworkLayer().NetworkFlow(), tcp, packet.Metadata().Timestamp)

		case <-ticker:
			// Every minute, flush connections that haven't seen activity in the past 2 minutes.
			flushed, closed := assembler.FlushOlderThan(time.Now().Add(time.Minute * -1))
			if flushed > 0 {
				glog.Infof("Flush : flushed=%d, closed=%d", flushed, closed)
			}
		}
	}
}

// httpStreamFactory implements tcpassembly.StreamFactory
type HttpStreamFactory struct {
	tracker *ConnectionTracker
}

func NewHttpStreamFactory(tr *ConnectionTracker) *HttpStreamFactory {
	return &HttpStreamFactory{tr}
}

func (h *HttpStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	assemblyStream := tcpreader.NewReaderStream()

	if transport.Src().String() == "80" {
		stream := NewSegmentedStream(&assemblyStream, nil, true)
		assemblyStream.LossErrors = true
		h.tracker.RegisterStreamLeg(net, transport, stream, Response)
		go util.SafeRun(func() { h.tracker.RunHttpResponseCycle(0, transport.Dst(), stream) })
	} else if transport.Dst().String() == "80" {
		stream := NewSegmentedStream(&assemblyStream, nil, true)
		assemblyStream.LossErrors = true
		h.tracker.RegisterStreamLeg(net, transport, stream, Request)
		go util.SafeRun(func() {
			h.tracker.RunHttpRequestCycle(0, transport.Src(), stream, fmt.Sprintf("%s:%s", net.Dst(), transport.Dst()))
		})
	} else {
		glog.Errorf("unknown transport %v", transport)
		go util.SafeRun(func() { tcpreader.DiscardBytesToFirstError(&assemblyStream) })
	}

	// ReaderStream implements tcpassembly.Stream, so we can return a pointer to it.
	return &assemblyStream
}
