package tlsproxy

import (
	"appscope.net/ca"
	"appscope.net/ews"
	"encoding/json"
	"github.com/golang/glog"
	"net/http"
)

type EwsListRequest struct {
	ClientIP  string
	AuthToken []byte
}

type EwsListResponse struct {
	Domains map[string]string
}

type EwsConfigureRequest struct {
	ClientIP  string
	AuthToken []byte
	Domains   map[string]string
}

func RegisterEws(pxy *TLSProxy) {
	pxy.exclByClientByDomain = make(map[string]map[string]string)
	
	ews.Register("POST", "/tls/v1/list_excl_domains", pxy.EwsList)
	ews.Register("POST", "/tls/v1/configure_excl_domains", pxy.EwsConfigure)
}

func (pxy *TLSProxy) EwsList(responseWriter http.ResponseWriter, request *http.Request) {
	var bytes []byte
	var ewsRequest EwsListRequest

	ews.SetCORSHeader(responseWriter)

	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(&ewsRequest)

	if err != nil {
		ews.RenderError(responseWriter, http.StatusInternalServerError, "E_WRONG_ARGS", err.Error())
		return
	}

	_, err = ca.Auth(ewsRequest.AuthToken)
	if err != nil {
		ews.RenderError(responseWriter, http.StatusUnauthorized, "E_CMD_FAIL", err.Error())
		return
	}

	glog.Infof("list, clientIP: %v", ewsRequest.ClientIP)

	var response EwsListResponse
	response.Domains = pxy.exclDomainsGet(ewsRequest.ClientIP)
	bytes, err = json.Marshal(response)

	if err != nil {
		ews.RenderError(responseWriter, http.StatusInternalServerError, "E_CMD_FAIL", err.Error())
	} else {
		responseWriter.Header().Set("Content-Type", "application/json")
		responseWriter.WriteHeader(http.StatusOK)
		responseWriter.Write(bytes)
	}
}

func (pxy *TLSProxy) EwsConfigure(responseWriter http.ResponseWriter, request *http.Request) {
	ews.SetCORSHeader(responseWriter)

	decoder := json.NewDecoder(request.Body)

	var ewsRequest EwsConfigureRequest
	err := decoder.Decode(&ewsRequest)

	if err != nil {
		ews.RenderError(responseWriter, http.StatusInternalServerError, "E_WRONG_ARGS", err.Error())
		return
	}

	_, err = ca.Auth(ewsRequest.AuthToken)
	if err != nil {
		ews.RenderError(responseWriter, http.StatusUnauthorized, "E_CMD_FAIL", err.Error())
		return
	}

	glog.Infof("configure, clientIP: %v", ewsRequest.ClientIP)

	pxy.exclDomainsSet(ewsRequest.ClientIP, ewsRequest.Domains)

	ews.RenderOk(responseWriter, "Domains set")
}
