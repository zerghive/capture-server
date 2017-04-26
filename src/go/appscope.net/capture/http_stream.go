package capture

import (
	"io"
	"net/http"
	"time"

	"appscope.net/event"
	"appscope.net/mem"

	"github.com/golang/glog"
	"github.com/google/gopacket"
)

const (
	cBodyLimit             = 1024 * 1024
	cMaxRingbufferLifetime = time.Minute * 20
)

func readSkip(r io.Reader, discardBuffer []byte) (discarded int, err error) {
	for {
		n, e := r.Read(discardBuffer)
		discarded += n
		if e != nil {
			return discarded, e
		}
	}
}

func (ct *ConnectionTracker) RunHttpRequestCycle(connId Id, client gopacket.Endpoint,
	cache *SegmentedStream, host string) {
	defer cache.Close()
	buf := cache.BufferedReader()

	discardBuffer := mem.GetBuffer()
	defer mem.RecycleBuffer(discardBuffer)

	for {
		id := cache.StartSegment(CSegmentedStreamNoLimit)
		if (cache.modifier != nil) {
			cache.modifier.SetRequestId(id)
		}
		
		req, err := http.ReadRequest(buf)
		t, d := cache.MarkDuration()

		if err != nil {
			glog.Warningf("%s Error parsing request for %v:", host, err)
			cache.ParseFailed(err)
			readSkip(cache, discardBuffer)
			
			ct.PostClientEvent(client, func() event.Event {
				return &HttpRequestEvent{
					HttpConnId:  connId,
					RequestId:   id,
					Info:        nil,
					Error:       err,
				}
			})
			
                        glog.Warningf("Request cycle exit, err = %v", err)
			return
		} else {
                        if (cache.modifier != nil) {
                                cache.modifier.RequestDst().MarkStart()
                        }
                    
			if host != "" && host != "" && req.Host != host {
				// glog.Warningf("Defined host was %s, got %s in request", host, req.Host)
			}
			if req.Host == "" {
				req.Host = host
			}
			if glog.V(vHttpRequests) {
				glog.Infof("[%v, %v] %s %s%s ", t, d, req.Method, req.Host, req.RequestURI)
			}
			
			payload := &RequestInfo{
				Host:             req.Host,
				RequestURI:       req.RequestURI,
				Header:           req.Header,
				Method:           req.Method,
				Proto:            req.Proto,
				TransferEncoding: req.TransferEncoding}
			cache.SetPayload(payload)

			ct.PostClientEvent(client, func() event.Event {
				return &HttpRequestEvent{
					HttpConnId:  connId,
					RequestId:   id,
					Info:        payload,
				}
			})

			cache.StartSegment(CSegmentedStreamDefaultLimit)
			if n, err := readSkip(req.Body, discardBuffer); err != io.EOF {
				glog.Warningf("Unexpected error reading request body after %d bytes: %v", n, err)
			}
			cache.MarkDuration()
                        if (cache.modifier != nil) {
                                cache.modifier.RequestDst().MarkEnd()
                        }
			req.Body.Close()
		}
	}
}

func (ct *ConnectionTracker) RunHttpResponseCycle(connId Id, client gopacket.Endpoint,
	cache *SegmentedStream) {
	defer cache.Close()
	buf := cache.BufferedReader()

	discardBuffer := mem.GetBuffer()
	defer mem.RecycleBuffer(discardBuffer)

	for {
		id := cache.StartSegment(CSegmentedStreamNoLimit)
		if (cache.modifier != nil) {
			cache.modifier.SetResponseId(id)
		}

		resp, err := http.ReadResponse(buf, nil)
		cache.MarkDuration()

		if err != nil {
			glog.Errorf("Error parsing response : %v", err)
			cache.ParseFailed(err)
			readSkip(cache, discardBuffer)
			
			ct.PostClientEvent(client, func() event.Event {
				return &HttpResponseEvent{
					HttpConnId: connId,
					RequestId:  id,
					Info:       nil,
					Error:      err,
				}
			})
			
                        glog.Warningf("Response cycle exit, err = %v", err)
			return
		} else {
			payload := &ResponseInfo{
				Header:           resp.Header,
				Code:             resp.StatusCode,
				Proto:            resp.Proto,
				TransferEncoding: resp.TransferEncoding}
			cache.SetPayload(payload)

			ct.PostClientEvent(client, func() event.Event {
				return &HttpResponseEvent{
					HttpConnId: connId,
					RequestId:  id,
					Info:       payload,
				}
			})

			cache.StartSegment(CSegmentedStreamDefaultLimit)

			if n, err := readSkip(resp.Body, discardBuffer); err != io.EOF {
				glog.Warning("Unexpected error reading response body after %d bytes : %v", n, err)
			}
			cache.MarkDuration()
			resp.Body.Close()
		}
	}
}

type ResponseInfo struct {
	Code  int    `json:"c"` // e.g. 200
	Proto string `json:"p"` // e.g. "HTTP/1.0"

	// Header maps header keys to values.  If the response had multiple
	// headers with the same key, they may be concatenated, with comma
	// delimiters.  (Section 4.2 of RFC 2616 requires that multiple headers
	// be semantically equivalent to a comma-delimited sequence.) Values
	// duplicated by other fields in this struct (e.g., ContentLength) are
	// omitted from Header.
	//
	// Keys in the map are canonicalized (see CanonicalHeaderKey).
	Header http.Header `json:"h"`

	// Contains transfer encodings from outer-most to inner-most. Value is
	// nil, means that "identity" encoding is used.
	TransferEncoding []string `json:"te",omitempty`
}

type RequestInfo struct {
	Proto string `json:"p"`

	Host string `json:"t"`

	// Method specifies the HTTP method (GET, POST, PUT, etc.).
	// For client requests an empty string means GET.
	Method string `json:"m"`

	// A header maps request lines to their values.
	// If the header says
	//
	//	accept-encoding: gzip, deflate
	//	Accept-Language: en-us
	//	Connection: keep-alive
	//
	// then
	//
	//	Header = map[string][]string{
	//		"Accept-Encoding": {"gzip, deflate"},
	//		"Accept-Language": {"en-us"},
	//		"Connection": {"keep-alive"},
	//	}
	//
	// HTTP defines that header names are case-insensitive.
	// The request parser implements this by canonicalizing the
	// name, making the first character and any characters
	// following a hyphen uppercase and the rest lowercase.
	//
	// For client requests certain headers are automatically
	// added and may override values in Header.
	//
	// See the documentation for the Request.Write method.
	Header http.Header `json:"h"`

	// TransferEncoding lists the transfer encodings from outermost to
	// innermost. An empty list denotes the "identity" encoding.
	// TransferEncoding can usually be ignored; chunked encoding is
	// automatically added and removed as necessary when sending and
	// receiving requests.
	TransferEncoding []string `json:"e",omitempty`

	// RequestURI is the unmodified Request-URI of the
	// Request-Line (RFC 2616, Section 5.1) as sent by the client
	// to a server. Usually the URL field should be used instead.
	// It is an error to set this field in an HTTP client request.
	RequestURI string `json:"r"`
}
