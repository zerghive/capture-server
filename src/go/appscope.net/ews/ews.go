package ews

import (
	"net/http"
	"sync"

	"appscope.net/util"
	"github.com/gorilla/mux"

	"github.com/golang/glog"
)

var lk sync.Mutex
var router *mux.Router = mux.NewRouter()

func init() {
	lk.Lock()
	defer lk.Unlock()

	router.Methods("OPTIONS").HandlerFunc(ewsOptionsRespond)
}

func Register(Method, Pattern string, HandlerFunc http.HandlerFunc) {
	lk.Lock()
	defer lk.Unlock()

	router.
		Methods(Method).
		Path(Pattern).
		Handler(HandlerFunc)
}

func SetCORSHeader(responseWriter http.ResponseWriter) {
	responseWriter.Header().Set("Access-Control-Allow-Origin", "*")
}

func ewsOptionsRespond(w http.ResponseWriter, r *http.Request) {
	SetCORSHeader(w)
	w.WriteHeader(http.StatusOK)
}

func log_req(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		glog.Infof("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func Run(listenTo string) {
	glog.Infof("Starting up embedded web server at %s", listenTo)
	go util.SafeRun(func() { glog.Errorf(http.ListenAndServe(listenTo, log_req(router)).Error()) })
}
