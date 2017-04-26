package capture

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"appscope.net/ca"
	"appscope.net/ews"

	"github.com/golang/glog"
	"github.com/gorilla/mux"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

const (
	cApiSecret = "65WsHzhE8ACmFVEtJ21G"
)

func RegisterEws(ct *ConnectionTracker) {
	ews.Register("POST", "/clients", ct.EwsClientIndex)
	ews.Register("GET", "/client/{clientIp}/eventstream", ct.EwsClientEventStream)
	ews.Register("POST", "/client/get_body", ct.EwsClientGetBody)
	ews.Register("POST", "/http/export", ct.EwsClientExport)
	ews.Register("POST", "/clients/connected", ct.EwsClientsConnected)
}

type ewsClientIndexRequest struct {
	ApiSecret string
}

type ewsClient struct {
	IP string
}

type ewsClientIndexResponse struct {
	Clients []ewsClient
}

type ewsGetBodyRequest struct {
	Type       string
	ClientIP   string
	HttpConnId Id `json:",string"`
	RequestId  int
	AuthToken  []byte
}

func (ct *ConnectionTracker) EwsClientIndex(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var req ewsClientIndexRequest
	if err := decoder.Decode(&req); err != nil {
		ews.RenderError(w, http.StatusBadRequest, "E_WRONG_ARGS", err.Error())
		return
	} else if req.ApiSecret != cApiSecret {
		ews.RenderError(w, http.StatusBadRequest, "E_NO_ACCESS", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	clients := []ewsClient{}
	for clientEndpoint, _ := range ct.clients {
		clients = append(clients, ewsClient{
			IP: fmt.Sprintf("%s", clientEndpoint),
		})
	}
	json.NewEncoder(w).Encode(clients)
}

type jConnInfo struct {
	Server, Port string
}

func (ct *ConnectionTracker) EwsTimeEcho(w http.ResponseWriter, r *http.Request) {
	tnow, err := time.Now().MarshalText()
	if err != nil {
		ews.RenderError(w, http.StatusInternalServerError, "Cannot marshall current time to text", "")
	} else {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write(tnow)
		glog.Infof("%v, %v", r.URL, string(tnow))
	}
}

var cEventTag = []byte("event: ")
var cDataTag = []byte("data: ")
var cNewline = []byte("\n")

func (ct *ConnectionTracker) EwsClientEventStream(w http.ResponseWriter, r *http.Request) {
	clientEndpoint := ip2endpoint(mux.Vars(r)["clientIp"])

	if clientEndpoint == nil {
		ews.RenderError(w, http.StatusNotFound, "E_NOT_FOUND", r.RemoteAddr)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	stream := ct.installEventStream(*clientEndpoint)
	defer ct.shutdownEventStream(*clientEndpoint)

	codec := json.NewEncoder(w)

	flush, _ := w.(http.Flusher)
	for data := range stream {

		w.Write(cEventTag)
		w.Write([]byte(data.Type()))
		w.Write(cNewline)

		w.Write(cDataTag)
		codec.Encode(data)

		if _, e := w.Write(cNewline); e != nil {
			glog.Errorf("Failed to encode %v : %v, shutting down event stream for %v",
				data, e, *clientEndpoint)
			return
		} else if flush != nil {
			flush.Flush()
		}
	}
}

func (ct *ConnectionTracker) EwsClientGetBody(w http.ResponseWriter, r *http.Request) {
	var ewsRequest ewsGetBodyRequest

	ews.SetCORSHeader(w)

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&ewsRequest)

	if err != nil {
		ews.RenderError(w, http.StatusInternalServerError, "E_WRONG_ARGS", err.Error())
		return
	}

	_, err = ca.Auth(ewsRequest.AuthToken)
	if err != nil {
		ews.RenderError(w, http.StatusUnauthorized, "E_CMD_FAIL", err.Error())
		return
	}

	glog.Infof("list, clientIP: %v", ewsRequest.ClientIP)

	clientInfo := ct.getClientInfo(*ip2endpoint(ewsRequest.ClientIP))

	var stream *SegmentedStream

	if connection, there := clientInfo.connectionByConnId[Id(ewsRequest.HttpConnId)]; there {
		switch ewsRequest.Type {
		case "request":
			stream = connection.requestStream
		case "response":
			stream = connection.responseStream
		}
	}

	if stream != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if _, _, err = stream.FlushSegment(ewsRequest.RequestId+1, w); err != nil {
			ews.RenderError(w, http.StatusInternalServerError, "E_NOT_FOUND", "RequestId")
		}
	} else {
		ews.RenderError(w, http.StatusInternalServerError, "E_NOT_FOUND", "HttpConnId")
	}
}

type ewsExportRequest struct {
	DeviceToken, ShareToken              []byte
	ActualStartEpochMs, ActualDurationMs int64 `json:",string"`
	DeviceId, OrgId                      string
}

func (ct *ConnectionTracker) EwsClientExport(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var req ewsExportRequest
	if err := decoder.Decode(&req); err != nil {
		ews.RenderError(w, http.StatusInternalServerError, "E_WRONG_ARGS", err.Error())
		return
	}

	var clientEndpoint *gopacket.Endpoint

	if len(req.DeviceId) > 0 && len(req.OrgId) > 0 {
		name := fmt.Sprintf("%s@%s", req.DeviceId, req.OrgId)
		ip, err := getDeviceIP(name)
		if err != nil {
			ews.RenderError(w, http.StatusNotFound, "E_NOT_FOUND", err.Error())
			return
		}
		clientEndpoint = ip2endpoint(ip)
	} else {
		clientEndpoint = ip2endpoint(strings.Split(r.RemoteAddr, ":")[0])
	}

	if clientEndpoint == nil {
		ews.RenderError(w, http.StatusNotFound, "E_NOT_FOUND", r.RemoteAddr)
		return
	}

	tmpDir, err := ct.Export(*clientEndpoint,
		req.ShareToken, req.DeviceToken,
		time.Unix(req.ActualStartEpochMs/1000, 0), time.Unix((req.ActualStartEpochMs+req.ActualDurationMs)/1000, 0))
	if err != nil {
		ews.RenderError(w, http.StatusInternalServerError, "E_CMD_FAIL", err.Error())
		return
	} else {
		json.NewEncoder(w).Encode(tmpDir)
	}
}

type ewsClientsConnectedRequest struct {
	AuthToken []byte
}

type ewsClientsConnectedResponse struct {
	Devices []VpnClient
}

func (ct *ConnectionTracker) EwsClientsConnected(w http.ResponseWriter, r *http.Request) {
	ews.SetCORSHeader(w)

	decoder := json.NewDecoder(r.Body)
	var req ewsClientsConnectedRequest
	if err := decoder.Decode(&req); err != nil {
		ews.RenderError(w, http.StatusInternalServerError, "E_WRONG_ARGS", err.Error())
		return
	}

	token, err := ca.Auth(req.AuthToken)
	if err != nil {
		ews.RenderError(w, http.StatusUnauthorized, "E_CMD_FAIL", err.Error())
		return
	}

	var resp ewsClientsConnectedResponse
	resp.Devices, err = GetOrganizationVpnClients(token.OrgId)
	if err != nil {
		ews.RenderError(w, http.StatusInternalServerError, "E_CMD_FAIL", err.Error())
		return
	}

	json.NewEncoder(w).Encode(resp)
}

func ip2endpoint(ip string) *gopacket.Endpoint {
	clientIp := net.ParseIP(ip)
	if clientIp == nil {
		return nil
	} else { // assuming we're only using IPv4 addresses for now
		clientIp = clientIp.To4()
	}

	ep := layers.NewIPEndpoint(clientIp)
	return &ep
}
