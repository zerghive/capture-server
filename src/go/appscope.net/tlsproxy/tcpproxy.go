package tlsproxy

import (
	"appscope.net/capture"
	"appscope.net/conntrac"
	"appscope.net/event"
	"appscope.net/filter"
	"appscope.net/util"

	"io"
	"net"
	"time"

	"github.com/golang/glog"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type TCPProxy struct {
	LocalAddr, RemoteAddr *string
	Tracker               *capture.ConnectionTracker
	FilterController      *filter.Controller
	EnableFiltering       bool
}

func (pxy *TCPProxy) Listen() error {
	addr, err := net.ResolveTCPAddr("tcp", *pxy.LocalAddr)
	if err != nil {
		glog.Errorf("Resolve local address error: %v", err)
		return err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		glog.Errorf("Listen %v error: %v", *pxy.LocalAddr, err)
		return err
	}

	glog.Infof("TCP HTTP Proxy enable filtering = %v", pxy.EnableFiltering)

	glog.Infof("TCP HTTP Proxy listening at %s", *pxy.LocalAddr)
	for {
		lConn, err := listener.AcceptTCP()
		if err == nil {
			go util.SafeRun(func() { pxy.connect(lConn) })
		} else {
			glog.Errorf("AcceptTCP() error %v", err)
		}
	}
}

func (pxy *TCPProxy) connect(lConn *net.TCPConn) {
	tm := time.Now()
	
	defer lConn.Close()
	
	connTrackRec := &conntrac.ConntracRec{
		Timestamp: uint64(time.Now().Unix()),
		ConnStartTime: &tm,
	}
	
	lConn, rAddr, serverName := GetOriginalDst(lConn, NoTlsServerNameLookup)
	connTrackRec.ClientIP = layers.NewIPEndpoint((lConn.RemoteAddr()).(*net.TCPAddr).IP.To4())
	connTrackRec.ServerName = serverName
	connTrackRec.ServerAddr = rAddr.String()
	
	if rAddr == nil {
		registerError(pxy.Tracker, connTrackRec, conntrac.SEGMENT_GET_ORIGINAL_DST_ERR)
		
		glog.Errorf("Failed to get original destination for %+v", lConn)
		return
	}

	rConn, err := net.DialTCP("tcp", nil, rAddr)
	if err != nil {
		registerError(pxy.Tracker, connTrackRec, conntrac.SEGMENT_RDIAL_TCP_ERR)
		
		glog.Errorf("Connect with remote host error %v", err)
		return
	}
	connTrackRec.ServerIP = layers.NewIPEndpoint((rConn.RemoteAddr()).(*net.TCPAddr).IP.To4())

	defer rConn.Close()

	if pxy.EnableFiltering && (pxy.FilterController != nil) {            
		modifier := pxy.FilterController.CreateModifier((lConn.RemoteAddr()).(*net.TCPAddr).IP.To4().String(),
			layers.NewIPEndpoint((lConn.RemoteAddr()).(*net.TCPAddr).IP.To4()),
			lConn, rConn, 
			pxy.Tracker)
                
                connectModifierAndRecord(pxy.Tracker, serverName, connTrackRec, lConn, rConn, modifier)
                
                pxy.FilterController.DeleteModifier(modifier)
	} else {
            connectAndRecord(pxy.Tracker, serverName, connTrackRec, lConn, rConn)
        }
        
        tm = time.Now()
        connTrackRec.ConnEndTime = &tm
}

func connectAndRecord(tracker *capture.ConnectionTracker, serverName string, conntracRec *conntrac.ConntracRec, lConn, rConn net.Conn) {

	var responseStream, requestStream *capture.SegmentedStream

	connectionId := tracker.NewId()
        
        requestStream = capture.NewSegmentedStream(lConn, rConn, false)
        responseStream = capture.NewSegmentedStream(rConn, lConn, false)

	tracker.PostClientEvent(conntracRec.ClientIP, func() event.Event {
		return &capture.HttpConnectionEvent{
			Id:    connectionId,
			Host:  serverName,
			State: capture.HttpConnectionStateRecording,
		}
	})

	tracker.RegisterStreamDuplex(connectionId, requestStream, responseStream, conntracRec)

	go util.SafeRun(func() {
		tracker.RunHttpRequestCycle(connectionId, conntracRec.ClientIP, requestStream, serverName)
		
                rConn.Close()
                lConn.Close()
                glog.Infof("Closed connections %v <> %v %s", conntracRec.ClientIP, conntracRec.ServerIP, serverName)
	})

	tracker.RunHttpResponseCycle(connectionId, conntracRec.ClientIP, responseStream)
        
        rConn.Close()
        lConn.Close()
        
        glog.Infof("Closed connections %v <> %v %s", conntracRec.ClientIP, conntracRec.ServerIP, serverName)
}

func connectModifierAndRecord(tracker *capture.ConnectionTracker, serverName string,
        conntracRec *conntrac.ConntracRec, lConn, rConn net.Conn, modifier *filter.Modifier) {

        var responseStream, requestStream *capture.SegmentedStream

        connectionId := tracker.NewId()

        requestStream = capture.NewSegmentedStream(modifier.RequestSrc(), modifier.RequestDst(), false)
        responseStream = capture.NewSegmentedStream(modifier.ResponseSrc(), modifier.ResponseDst(), false)
        
        requestStream.RegisterModifier(modifier)
	responseStream.RegisterModifier(modifier)
                        
        go util.SafeRun(func() { 
            modifier.RunRequestCycle(uint64(connectionId))
            
            rConn.Close()
            lConn.Close()
            glog.Infof("Closed connections %v <> %v %s", conntracRec.ClientIP, conntracRec.ServerIP, serverName)
        })
        
        go util.SafeRun(func() { 
            modifier.RunResponseCycle(uint64(connectionId))
            
            rConn.Close()
            lConn.Close()
            glog.Infof("Closed connections %v <> %v %s", conntracRec.ClientIP, conntracRec.ServerIP, serverName)
        })

        tracker.PostClientEvent(conntracRec.ClientIP, func() event.Event {
                return &capture.HttpConnectionEvent{
                        Id:    connectionId,
                        Host:  serverName,
                        State: capture.HttpConnectionStateRecording,
                }
        })

        tracker.RegisterStreamDuplex(connectionId, requestStream, responseStream, conntracRec)

        go util.SafeRun(func() {
                tracker.RunHttpRequestCycle(connectionId, conntracRec.ClientIP, requestStream, serverName)
        })

        tracker.RunHttpResponseCycle(connectionId, conntracRec.ClientIP, responseStream)
}

func connectNoRecord(tracker *capture.ConnectionTracker, serverName string, clientTcpAddr, serverTcpAddr gopacket.Endpoint, lConn, rConn net.Conn) {
	done := make(chan bool)

	go util.SafeRun(func() {
		io.Copy(rConn, lConn)
		lConn.Close()
		rConn.Close()
		done <- true
	})

	io.Copy(lConn, rConn)
	rConn.Close()
	lConn.Close()

	glog.Infof("Closed connections %v <> %v %s", clientTcpAddr, serverTcpAddr, serverName)
	<-done
}

func connectNoRecording(lConn net.Conn, rConn net.Conn) {
	if glog.V(vConnectionTrace) {
		glog.Infof("pipe %v %v", lConn.RemoteAddr(), rConn.RemoteAddr())
	}

	done := make(chan bool)
	go util.SafeRun(func() {
		io.Copy(lConn, rConn)
		shut := false
	retry_1:
		select {
		case done <- true:
			return
		case <-time.After(transparentTimeout):
			if shut == false {
				glog.Warningf("Shutting down client connection %s as server is gone", lConn.RemoteAddr())
				lConn.Close()
				shut = true
			} else {
				glog.Errorf("Still waiting for %s to exhaust", lConn.RemoteAddr())
			}
			goto retry_1
		}
	})

	io.Copy(rConn, lConn)
	shut := false
retry_2:
	select {
	case <-done:
		return
	case <-time.After(transparentTimeout):
		if shut == false {
			glog.Warningf("Shutting down server connection %s as client is gone", rConn.RemoteAddr())
			rConn.Close()
			shut = true
		} else {
			glog.Errorf("Still waiting for %s to exhaust", rConn.RemoteAddr())
		}
		goto retry_2
	}
}

func registerError(tracker *capture.ConnectionTracker, connTrackRec *conntrac.ConntracRec, err conntrac.SegmentError) {
	connectionId := tracker.NewId()
	
        connTrackRec.ConnErr = err
        tm := time.Now()
        connTrackRec.ConnEndTime = &tm
	
	tracker.RegisterStreamDuplex(connectionId, nil, nil, connTrackRec)
	
	tracker.PostClientEvent(connTrackRec.ClientIP, func() event.Event {
                return &capture.ConnectionErrorEvent{
			Id:         connectionId,
			Error:      connTrackRec.ConnErr,
			Tls:        connTrackRec.TlsConn,
			ServerName: connTrackRec.ServerName,
			ServerAddr: connTrackRec.ServerAddr,
                }
        })
}
