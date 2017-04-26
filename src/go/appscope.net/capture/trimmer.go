package capture

import (
	"container/list"
	"time"

	"appscope.net/mem"
	"github.com/golang/glog"
)

/*
 launched by connection tracker
 walks over connections to see whether we need to trim some of the data in them
 relies on client-reported stored buffer
*/

func (ct *ConnectionTracker) doTrimming(v interface{}) (err error) {
	ct.Lock()
	defer ct.Unlock()

	tStarted := time.Now()

	// TODO : move on to client-defined timestamp
	cutOffTimeClosedConn := tStarted.Add(-time.Minute * 30)
	cutOffTimeHttpErrConn := tStarted.Add(-time.Minute * 30)
	cutOffTimeFailedConn := tStarted.Add(-time.Minute * 60)
	
	stillOpen := 0
	swept := 0
	for _, clientInfo := range ct.clients {
		toDelete := make([]*list.Element, 0, 10)
		for conn := clientInfo.connections.Front(); conn != nil; conn = conn.Next() {
			connection := conn.Value.(*connection)
			if connection.requestStream == nil || connection.responseStream == nil {
				if connection.natRec.ConnEndTime.Before(cutOffTimeFailedConn) {
					if connection.requestStream != nil {
						connection.requestStream.Destroy()
					}
					if connection.responseStream != nil {
						connection.responseStream.Destroy()
					}
					
					toDelete = append(toDelete, conn)
					swept++
				} else {
					stillOpen++
					if glog.V(vConnectionTrimmer) {
						glog.Infof("Still open (req=%v, resp=%v): %v", connection.requestStream,
							connection.responseStream, connection.natRec)
					}
				}
			} else {
				if connection.requestStream.closed == true && connection.responseStream.closed == true {
					var cutOffTime time.Time
					if connection.requestStream.IsFailed() || connection.responseStream.IsFailed() {
						cutOffTime = cutOffTimeHttpErrConn
					} else {
						cutOffTime = cutOffTimeClosedConn
					}
					
					if _, e1 := connection.responseStream.GetTimeRange(); e1 != nil && e1.Before(cutOffTime) {
						connection.requestStream.Destroy()
						connection.responseStream.Destroy()
						toDelete = append(toDelete, conn)
						swept++
						//glog.Infof("Swept %v", connection.natRec)
					}
				} else {
					stillOpen++
					if glog.V(vConnectionTrimmer) {
						glog.Infof("Still open (req=%v, resp=%v): %v", connection.requestStream.closed,
							connection.responseStream.closed, connection.natRec)
					}
				}
			}
		}

		for _, e := range toDelete {
			delete(clientInfo.connectionByConnId, (e.Value.(*connection)).connId)
			clientInfo.connections.Remove(e)
		}
	}
	if glog.V(vConnectionTrimmer) {
		bufOut := mem.GetAllocatedBuffersCount()
		glog.Infof("Trimmer done. Clients=%d, open=%d, swept=%d, buffers out=%d", len(ct.clients), stillOpen, swept, bufOut)
		if bufOut > 0 && stillOpen == 0 {
			// count buffers which are in PENDING
			for c, rqrsp := range ct.clients_pending {
				var bfpending = 0
				if rqrsp != nil {
					if rqrsp[0] != nil {
						bfpending += len(rqrsp[0].streamData.chunks)
					}
					if rqrsp[1] != nil {
						bfpending += len(rqrsp[1].streamData.chunks)
					}
				}
				if bfpending > 0 {
					glog.Infof(" in pending %v : %v", c, bfpending)
				}
			}
		}
	}
	return nil
}
