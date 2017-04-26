package capture

import (
	"appscope.net/conntrac"
	"appscope.net/upload"

	"github.com/golang/glog"
	"github.com/google/gopacket"

	"container/list"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

var TlsCipherCodes = map[uint16]uint16 {
    tls.TLS_RSA_WITH_RC4_128_SHA:                1,
    tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA:           2,
    tls.TLS_RSA_WITH_AES_128_CBC_SHA:            3,
    tls.TLS_RSA_WITH_AES_256_CBC_SHA:            4,
    tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA:        5,
    tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA:    6,
    tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:    7,
    tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA:          8,
    tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA:     9,
    tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA:      10,
    tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:      11,
    tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:   12,
    tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256: 13,
    tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:   14,
    tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384: 15,
    tls.TLS_FALLBACK_SCSV:                       16,
}

var TlsVersionCodes = map[uint16]uint16 {
    tls.VersionSSL30: 1,
    tls.VersionTLS10: 2,
    tls.VersionTLS11: 3,
    tls.VersionTLS12: 4,
}

func (ct *ConnectionTracker) Export(client gopacket.Endpoint,
	shareToken, deviceToken []byte,
	from, to time.Time) (string, error) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "proxy-")
	if err != nil {
		fmt.Printf("Mkdir %s failed: %v", tmpDir, err)
		return "", err
	} else {
		glog.Infof("Exporting %s [%s:%s] to %s\n", client, from, to, tmpDir)
	}

	ct.debugDump()

	ct.qExport.Push(&cmdExport{client, shareToken, deviceToken, from, to, tmpDir})

	return tmpDir, nil
}

type cmdExport struct {
	client                  gopacket.Endpoint
	shareToken, deviceToken []byte
	from, to                time.Time
	tmpDir                  string
}

type jobExport struct {
	cmd         *cmdExport
	connections *list.List
	serial      Id
}

type exportData struct {
	Connections []ConnectionInfo
	Requests []HttpReqResp
	From, To time.Time
}

func (ct *ConnectionTracker) doExportCmd(v interface{}) (err error) {
	cmd := v.(*cmdExport)

	if ct.clients[cmd.client] == nil {
		err = fmt.Errorf("Nothing to export for %v", cmd.client)
		return
	}

	ct.qExportJobs.Push(&jobExport{
		cmd:         cmd,
		connections: ct.clients[cmd.client].connections,
		serial:      ct.clients[cmd.client].serial,
	})

	ct.clients[cmd.client].connections = list.New()
	return nil
}

func (ct *ConnectionTracker) doExportJob(v interface{}) (err error) {
	// return data back to tracker when done
	defer ct.qExportComplete.Push(v)

	job := v.(*jobExport)
	exportLogFile, err := os.Create(path.Join(job.cmd.tmpDir, "export.log"))
	if err != nil {
		err = fmt.Errorf("Failed create export.log file for job %s : %v", job.cmd.tmpDir, err)
		return
	}
	defer exportLogFile.Close()

	connections := make([]ConnectionInfo, 0, job.connections.Len())
	requests := make([]HttpReqResp, 0, job.connections.Len())
	fileInfo := make([]upload.FileInfo, 0, job.connections.Len())

	exportLog := log.New(exportLogFile, "", log.Lshortfile)
	fileInfo = append(fileInfo, upload.FileInfo{Name: "export.log", MimeType: "text/plain"})

	file, err := os.Create(path.Join(job.cmd.tmpDir, "data"))
	if err != nil {
		return err
	}
	defer file.Close()
	fileInfo = append(fileInfo, upload.FileInfo{
		Name:     "data",
		MimeType: "application/octet-stream"})

	dataOffset := 0
	seqNo := 0
	for conn := job.connections.Front(); conn != nil; conn = conn.Next() {
		connectionInfo := ConnectionInfo {
			ConnectionId:    seqNo,
			ConnectionStart: conn.Value.(*connection).natRec.ConnStartTime,
			ConnectionEnd:   conn.Value.(*connection).natRec.ConnEndTime,
			ServerName:      conn.Value.(*connection).natRec.ServerName,
			ServerAddr:      conn.Value.(*connection).natRec.ServerAddr,
			Error:           conn.Value.(*connection).natRec.ConnErr,
		}
		if conn.Value.(*connection).natRec.TlsConn {
			connectionInfo.ClientTls = &TlsInfo {
				HandshakeStart: conn.Value.(*connection).natRec.ClientTlsHandshakeStart,
				HandshakeDuration: conn.Value.(*connection).natRec.ClientTlsHandshakeDuration,
				Version: TlsVersionCodes[conn.Value.(*connection).natRec.ClientTlsVersion],
				Cipher: TlsCipherCodes[conn.Value.(*connection).natRec.ClientTlsCipher],
				Proto: conn.Value.(*connection).natRec.ClientTlsProto,
			}
			connectionInfo.ServerTls = &TlsInfo {
				HandshakeStart: conn.Value.(*connection).natRec.ServerTlsHandshakeStart,
				HandshakeDuration: conn.Value.(*connection).natRec.ServerTlsHandshakeDuration,
				Version: TlsVersionCodes[conn.Value.(*connection).natRec.ServerTlsVersion],
				Cipher: TlsCipherCodes[conn.Value.(*connection).natRec.ServerTlsCipher],
				Proto: conn.Value.(*connection).natRec.ServerTlsProto,
			}
		}
		connections = append(connections, connectionInfo)
		
		if conn.Value.(*connection).requestStream != nil && conn.Value.(*connection).responseStream != nil {
			rqrsp, n, err := dumpReqRespStream(exportLog, conn.Value.(*connection), seqNo,
				job.cmd.from, job.cmd.to, file, dataOffset)
			if err != nil {
				glog.Errorf(err.Error())
				exportLog.Println(err)
			}
			dataOffset = n
			if len(rqrsp) > 0 {
				requests = append(requests, rqrsp...)
			}
		}
		seqNo++
	}
	file.Close()

	file, err = os.Create(path.Join(job.cmd.tmpDir, "http.json"))
	if err != nil {
		log.Println(err)
		return
	}
	fileInfo = append(fileInfo, upload.FileInfo{Name: "http.json", MimeType: "application/json"})

	data := exportData{
		Connections: connections,
		Requests:    requests,
		From:        job.cmd.from,
		To:          job.cmd.to,
	}
	json.NewEncoder(file).Encode(data)
	file.Close()

	// kick off upload
	upload.GetService().Add(job.cmd.tmpDir, fileInfo, job.cmd.deviceToken, job.cmd.shareToken)

	return nil
}

func (ct *ConnectionTracker) doExportComplete(v interface{}) (err error) {
	job := v.(*jobExport)

	if clEx := ct.clients[job.cmd.client]; clEx == nil {
		glog.Infof("Cannot return back connection data after export for %v, no record for it (disconnected?)", job.cmd.client)
		discardConnections(job.connections)
	} else if clEx.serial != job.serial {
		// means its another client, we won't return data to him
		glog.Info("Serial don't match for %v : have %d, want %d - discarding data", job.cmd.client, job.serial, clEx.serial)
		discardConnections(job.connections)
	} else {
		ct.clients[job.cmd.client].connections.PushBackList(job.connections)
	}
	return
}

func discardConnections(conns *list.List) {
	panic("not implemented")
}

type SegmentInfo struct {
	Received        time.Time     `json:"t"`
	Duration        time.Duration `json:"d"`
	Offset          int           `json:"o"`
	Length          int           `json:"l"`
	ProcessedLength int           `json:"p"`
}

type TlsInfo struct {
	HandshakeStart    *time.Time    `json:"t"`
	HandshakeDuration time.Duration `json:"d"`
	Version           uint16        `json:"v"`
	Cipher            uint16        `json:"c"`
	Proto             string        `json:"p"`
}

type ConnectionInfo struct {
	ConnectionId     int        `json:"c"`
	ConnectionStart  *time.Time `json:"t"`
	ConnectionEnd    *time.Time `json:"f"`
	ServerName       string     `json:"n"`
	ServerAddr       string     `json:"a"`
	Error conntrac.SegmentError `json:"e"`
	ClientTls        *TlsInfo   `json:"tlsc"`
	ServerTls        *TlsInfo   `json:"tlss"`
}

const (
	SegmentRequestHeader  = 0
	SegmentRequestBody    = 1
	SegmentResponseHeader = 2
	SegmentResponseBody   = 3
)

type HttpReqResp struct {
	ConnectionId int `json:"c"`

	RequestInfo  *RequestInfo   `json:"rq"`
	ResponseInfo *ResponseInfo  `json:"rs"`

	Segments [4]SegmentInfo `json:"s"`
}

func dumpReqRespStream(lg *log.Logger, conn *connection, connID int, from, to time.Time, out io.Writer, initialOffset int) ([]HttpReqResp, int, error) {
	// determine segments which are falling this time limit

	segmentIds := conn.requestStream.GetEvenSegments(from, to)
	reqResp := make([]HttpReqResp, len(segmentIds), len(segmentIds))
		
	offset := initialOffset

	for reqNo, id := range segmentIds {		
		reqPayload, _ := conn.requestStream.GetPayload(id)
		respPayload, _ := conn.responseStream.GetPayload(id)
		
		reqResp[reqNo].ConnectionId = connID
		if reqPayload != nil {
			reqResp[reqNo].RequestInfo = reqPayload.(*RequestInfo)
		}
		if respPayload != nil {
			reqResp[reqNo].ResponseInfo = respPayload.(*ResponseInfo)
		}
		
		if reqPayload != nil {
			storeSegment(&reqResp[reqNo].Segments[SegmentRequestHeader], id, out, conn.requestStream.FlushSegment)
			storeSegment(&reqResp[reqNo].Segments[SegmentRequestBody], id+1, out, conn.requestStream.FlushSegment)
		}
		if respPayload != nil {
			storeSegment(&reqResp[reqNo].Segments[SegmentResponseHeader], id, out, conn.responseStream.FlushSegment)
			storeSegment(&reqResp[reqNo].Segments[SegmentResponseBody], id+1, out, conn.responseStream.FlushSegment)
		}

		for i, _ := range reqResp[reqNo].Segments {
			reqResp[reqNo].Segments[i].Offset = offset
			offset += reqResp[reqNo].Segments[i].Length
		}

		reqResp[reqNo].Segments[SegmentRequestHeader].Received,
			reqResp[reqNo].Segments[SegmentRequestHeader].Duration = conn.requestStream.GetSegmentTime(id)
		reqResp[reqNo].Segments[SegmentRequestBody].Received,
			reqResp[reqNo].Segments[SegmentRequestBody].Duration = conn.requestStream.GetSegmentTime(id + 1)
		
		reqResp[reqNo].Segments[SegmentResponseHeader].Received,
			reqResp[reqNo].Segments[SegmentResponseHeader].Duration = conn.responseStream.GetSegmentTime(id)
		reqResp[reqNo].Segments[SegmentResponseBody].Received,
			reqResp[reqNo].Segments[SegmentResponseBody].Duration = conn.responseStream.GetSegmentTime(id + 1)
	}

	return reqResp, offset, nil
}

func storeSegment(segment *SegmentInfo, id int, out io.Writer,
	save func(id int, out io.Writer) (int, int, error)) (err error) {

	segment.Length, segment.ProcessedLength, err = save(id, out)
	return err
}
