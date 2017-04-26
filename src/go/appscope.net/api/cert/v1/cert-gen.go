// Package cert provides access to the .
//
// Usage example:
//
//   import "appscope.net/api/cert/v1"
//   ...
//   certService, err := cert.New(oauthHttpClient)
package cert // import "appscope.net/api/cert/v1"

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	context "golang.org/x/net/context"
	ctxhttp "golang.org/x/net/context/ctxhttp"
	gensupport "google.golang.org/api/gensupport"
	googleapi "google.golang.org/api/googleapi"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Always reference these packages, just in case the auto-generated code
// below doesn't.
var _ = bytes.NewBuffer
var _ = strconv.Itoa
var _ = fmt.Sprintf
var _ = json.NewDecoder
var _ = io.Copy
var _ = url.Parse
var _ = gensupport.MarshalJSON
var _ = googleapi.Version
var _ = errors.New
var _ = strings.Replace
var _ = context.Canceled
var _ = ctxhttp.Do

const apiId = "cert:v1"
const apiName = "cert"
const apiVersion = "v1"
const basePath = "https://appnetscope.appspot.com/_ah/api/cert/v1/"

func New(client *http.Client) (*Service, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	s := &Service{client: client, BasePath: basePath}
	return s, nil
}

type Service struct {
	client    *http.Client
	BasePath  string // API endpoint base URL
	UserAgent string // optional additional User-Agent fragment
}

func (s *Service) userAgent() string {
	if s.UserAgent == "" {
		return googleapi.UserAgent
	}
	return googleapi.UserAgent + " " + s.UserAgent
}

type ApiError struct {
	Code string `json:"code,omitempty"`

	Message string `json:"message,omitempty"`

	Retry bool `json:"retry,omitempty"`

	Url string `json:"url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Code") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ApiError) MarshalJSON() ([]byte, error) {
	type noMethod ApiError
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type SignAndroidCertRequest struct {
	DeviceToken []byte `json:"DeviceToken,omitempty"`

	HardwareClass string `json:"HardwareClass,omitempty"`

	HardwareModel string `json:"HardwareModel,omitempty"`

	Name string `json:"Name,omitempty"`

	Udid string `json:"Udid,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DeviceToken") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *SignAndroidCertRequest) MarshalJSON() ([]byte, error) {
	type noMethod SignAndroidCertRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type SignAndroidCertResponse struct {
	CA []byte `json:"CA,omitempty"`

	CAUrl string `json:"CAUrl,omitempty"`

	Cert []byte `json:"Cert,omitempty"`

	Error *ApiError `json:"Error,omitempty"`

	Key []byte `json:"Key,omitempty"`

	NotAfter string `json:"NotAfter,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "CA") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *SignAndroidCertResponse) MarshalJSON() ([]byte, error) {
	type noMethod SignAndroidCertResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type SignCertResponse struct {
	CA []byte `json:"CA,omitempty"`

	CAUrl string `json:"CAUrl,omitempty"`

	Cert []byte `json:"Cert,omitempty"`

	Error *ApiError `json:"Error,omitempty"`

	NotAfter string `json:"NotAfter,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "CA") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *SignCertResponse) MarshalJSON() ([]byte, error) {
	type noMethod SignCertResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type SignClientCertRequest struct {
	DeviceToken []byte `json:"DeviceToken,omitempty"`

	PublicKey []byte `json:"PublicKey,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DeviceToken") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *SignClientCertRequest) MarshalJSON() ([]byte, error) {
	type noMethod SignClientCertRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type SignVPNCertRequest struct {
	PublicKey []byte `json:"PublicKey,omitempty"`

	// ForceSendFields is a list of field names (e.g. "PublicKey") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *SignVPNCertRequest) MarshalJSON() ([]byte, error) {
	type noMethod SignVPNCertRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type SignVPNCertResponse struct {
	Error *ApiError `json:"Error,omitempty"`

	RootCA []byte `json:"RootCA,omitempty"`

	RootCAUrl string `json:"RootCAUrl,omitempty"`

	TlsCA []byte `json:"TlsCA,omitempty"`

	TlsCAKey []byte `json:"TlsCAKey,omitempty"`

	TlsCAUrl string `json:"TlsCAUrl,omitempty"`

	VpnCert []byte `json:"VpnCert,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Error") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *SignVPNCertResponse) MarshalJSON() ([]byte, error) {
	type noMethod SignVPNCertResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// method id "cert.signandroidcert":

type SignandroidcertCall struct {
	s                      *Service
	signandroidcertrequest *SignAndroidCertRequest
	urlParams_             gensupport.URLParams
	ctx_                   context.Context
}

// Signandroidcert:
func (s *Service) Signandroidcert(signandroidcertrequest *SignAndroidCertRequest) *SignandroidcertCall {
	c := &SignandroidcertCall{s: s, urlParams_: make(gensupport.URLParams)}
	c.signandroidcertrequest = signandroidcertrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *SignandroidcertCall) Fields(s ...googleapi.Field) *SignandroidcertCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *SignandroidcertCall) Context(ctx context.Context) *SignandroidcertCall {
	c.ctx_ = ctx
	return c
}

func (c *SignandroidcertCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.signandroidcertrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "signandroidcert")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "cert.signandroidcert" call.
// Exactly one of *SignAndroidCertResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *SignAndroidCertResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *SignandroidcertCall) Do() (*SignAndroidCertResponse, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &SignAndroidCertResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "httpMethod": "POST",
	//   "id": "cert.signandroidcert",
	//   "path": "signandroidcert",
	//   "request": {
	//     "$ref": "SignAndroidCertRequest",
	//     "parameterName": "resource"
	//   },
	//   "response": {
	//     "$ref": "SignAndroidCertResponse"
	//   }
	// }

}

// method id "cert.signclientcert":

type SignclientcertCall struct {
	s                     *Service
	signclientcertrequest *SignClientCertRequest
	urlParams_            gensupport.URLParams
	ctx_                  context.Context
}

// Signclientcert:
func (s *Service) Signclientcert(signclientcertrequest *SignClientCertRequest) *SignclientcertCall {
	c := &SignclientcertCall{s: s, urlParams_: make(gensupport.URLParams)}
	c.signclientcertrequest = signclientcertrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *SignclientcertCall) Fields(s ...googleapi.Field) *SignclientcertCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *SignclientcertCall) Context(ctx context.Context) *SignclientcertCall {
	c.ctx_ = ctx
	return c
}

func (c *SignclientcertCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.signclientcertrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "signclientcert")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "cert.signclientcert" call.
// Exactly one of *SignCertResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *SignCertResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *SignclientcertCall) Do() (*SignCertResponse, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &SignCertResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "httpMethod": "POST",
	//   "id": "cert.signclientcert",
	//   "path": "signclientcert",
	//   "request": {
	//     "$ref": "SignClientCertRequest",
	//     "parameterName": "resource"
	//   },
	//   "response": {
	//     "$ref": "SignCertResponse"
	//   }
	// }

}

// method id "cert.signvpncert":

type SignvpncertCall struct {
	s                  *Service
	signvpncertrequest *SignVPNCertRequest
	urlParams_         gensupport.URLParams
	ctx_               context.Context
}

// Signvpncert:
func (s *Service) Signvpncert(signvpncertrequest *SignVPNCertRequest) *SignvpncertCall {
	c := &SignvpncertCall{s: s, urlParams_: make(gensupport.URLParams)}
	c.signvpncertrequest = signvpncertrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *SignvpncertCall) Fields(s ...googleapi.Field) *SignvpncertCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *SignvpncertCall) Context(ctx context.Context) *SignvpncertCall {
	c.ctx_ = ctx
	return c
}

func (c *SignvpncertCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.signvpncertrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "signvpncert")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "cert.signvpncert" call.
// Exactly one of *SignVPNCertResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *SignVPNCertResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *SignvpncertCall) Do() (*SignVPNCertResponse, error) {
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &SignVPNCertResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "httpMethod": "POST",
	//   "id": "cert.signvpncert",
	//   "path": "signvpncert",
	//   "request": {
	//     "$ref": "SignVPNCertRequest",
	//     "parameterName": "resource"
	//   },
	//   "response": {
	//     "$ref": "SignVPNCertResponse"
	//   }
	// }

}
