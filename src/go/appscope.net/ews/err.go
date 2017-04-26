package ews

import (
	"encoding/json"
	"net/http"

	"github.com/golang/glog"
)

type Err struct {
	Code, Text string
}

func RenderError(w http.ResponseWriter, httpCode int, errCode, reason string) {
	glog.Errorf("%s %s", errCode, reason)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(httpCode)
	if err := json.NewEncoder(w).Encode(Err{Code: errCode, Text: reason}); err != nil {
	}
}

func RenderOk(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Err{Code: "OK", Text: message})
}
