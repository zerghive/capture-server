// Package share provides access to the .
//
// Usage example:
//
//   import "appscope.net/api/share/v1"
//   ...
//   shareService, err := share.New(oauthHttpClient)
package share // import "appscope.net/api/share/v1"

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

const apiId = "share:v1"
const apiName = "share"
const apiVersion = "v1"
const basePath = "https://appnetscope.appspot.com/_ah/api/share/v1/"

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

type AppendFilesRequest struct {
	Files []*FileItem `json:"Files,omitempty"`

	ShareToken []byte `json:"ShareToken,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Files") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *AppendFilesRequest) MarshalJSON() ([]byte, error) {
	type noMethod AppendFilesRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type AppendNotesRequest struct {
	Notes []*Note `json:"Notes,omitempty"`

	ShareToken []byte `json:"ShareToken,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Notes") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *AppendNotesRequest) MarshalJSON() ([]byte, error) {
	type noMethod AppendNotesRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type ConfirmUploadRequest struct {
	DeviceToken []byte `json:"DeviceToken,omitempty"`

	UploadTokens []*Token `json:"UploadTokens,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DeviceToken") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ConfirmUploadRequest) MarshalJSON() ([]byte, error) {
	type noMethod ConfirmUploadRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type CreateReq struct {
	Description string `json:"Description,omitempty"`

	DeviceToken []byte `json:"DeviceToken,omitempty"`

	Email string `json:"Email,omitempty"`

	Title string `json:"Title,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Description") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CreateReq) MarshalJSON() ([]byte, error) {
	type noMethod CreateReq
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type CreateResp struct {
	Error *ApiError `json:"Error,omitempty"`

	ShareToken []byte `json:"ShareToken,omitempty"`

	ShareUrl string `json:"ShareUrl,omitempty"`

	UsePrivateStorage bool `json:"UsePrivateStorage,omitempty"`

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

func (s *CreateResp) MarshalJSON() ([]byte, error) {
	type noMethod CreateResp
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type DeleteSnapshotRequest struct {
	DeviceToken []byte `json:"DeviceToken,omitempty"`

	ShareToken []byte `json:"ShareToken,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DeviceToken") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *DeleteSnapshotRequest) MarshalJSON() ([]byte, error) {
	type noMethod DeleteSnapshotRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type DownloadFileItem struct {
	Name string `json:"Name,omitempty"`

	Status string `json:"Status,omitempty"`

	Tag string `json:"Tag,omitempty"`

	Url string `json:"Url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Name") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *DownloadFileItem) MarshalJSON() ([]byte, error) {
	type noMethod DownloadFileItem
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type FileAccessRequest struct {
	ShareToken []byte `json:"ShareToken,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ShareToken") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *FileAccessRequest) MarshalJSON() ([]byte, error) {
	type noMethod FileAccessRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type FileAccessResponse struct {
	CreatedOn string `json:"CreatedOn,omitempty"`

	Description string `json:"Description,omitempty"`

	Error *ApiError `json:"Error,omitempty"`

	Files []*DownloadFileItem `json:"Files,omitempty"`

	Notes []*Note `json:"Notes,omitempty"`

	Title string `json:"Title,omitempty"`

	UsePrivateStorage bool `json:"UsePrivateStorage,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "CreatedOn") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *FileAccessResponse) MarshalJSON() ([]byte, error) {
	type noMethod FileAccessResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type FileGroupInfo struct {
	Status string `json:"Status,omitempty"`

	Tag string `json:"Tag,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Status") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *FileGroupInfo) MarshalJSON() ([]byte, error) {
	type noMethod FileGroupInfo
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type FileItem struct {
	MimeType string `json:"MimeType,omitempty"`

	Name string `json:"Name,omitempty"`

	Size int64 `json:"Size,omitempty"`

	Tag string `json:"Tag,omitempty"`

	// ForceSendFields is a list of field names (e.g. "MimeType") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *FileItem) MarshalJSON() ([]byte, error) {
	type noMethod FileItem
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type GetSharesRequest struct {
	DeviceToken []byte `json:"DeviceToken,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DeviceToken") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *GetSharesRequest) MarshalJSON() ([]byte, error) {
	type noMethod GetSharesRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type GetSharesResponse struct {
	Error *ApiError `json:"Error,omitempty"`

	Items []*ShareItem `json:"Items,omitempty"`

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

func (s *GetSharesResponse) MarshalJSON() ([]byte, error) {
	type noMethod GetSharesResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type Note struct {
	Data string `json:"Data,omitempty"`

	Id string `json:"Id,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Data") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Note) MarshalJSON() ([]byte, error) {
	type noMethod Note
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type ShareItem struct {
	Description string `json:"Description,omitempty"`

	FileGroupStatus []*FileGroupInfo `json:"FileGroupStatus,omitempty"`

	Title string `json:"Title,omitempty"`

	UploadedOn string `json:"UploadedOn,omitempty"`

	Url string `json:"Url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Description") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ShareItem) MarshalJSON() ([]byte, error) {
	type noMethod ShareItem
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type Token struct {
	Token []byte `json:"Token,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Token") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Token) MarshalJSON() ([]byte, error) {
	type noMethod Token
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type UploadFileItem struct {
	MimeType string `json:"MimeType,omitempty"`

	Name string `json:"Name,omitempty"`

	Size int64 `json:"Size,omitempty"`

	Tag string `json:"Tag,omitempty"`

	Token []byte `json:"Token,omitempty"`

	Url string `json:"Url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "MimeType") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *UploadFileItem) MarshalJSON() ([]byte, error) {
	type noMethod UploadFileItem
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type UploadFilesRequest struct {
	DeviceToken []byte `json:"DeviceToken,omitempty"`

	Files []*FileItem `json:"Files,omitempty"`

	ShareToken []byte `json:"ShareToken,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DeviceToken") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *UploadFilesRequest) MarshalJSON() ([]byte, error) {
	type noMethod UploadFilesRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type UploadFilesResponse struct {
	Error *ApiError `json:"Error,omitempty"`

	Files []*UploadFileItem `json:"Files,omitempty"`

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

func (s *UploadFilesResponse) MarshalJSON() ([]byte, error) {
	type noMethod UploadFilesResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type UploadStatusRequest struct {
	DeviceToken []byte `json:"DeviceToken,omitempty"`

	Status string `json:"Status,omitempty"`

	UploadToken []byte `json:"UploadToken,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DeviceToken") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *UploadStatusRequest) MarshalJSON() ([]byte, error) {
	type noMethod UploadStatusRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type VoidResponse struct {
	Error *ApiError `json:"Error,omitempty"`

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

func (s *VoidResponse) MarshalJSON() ([]byte, error) {
	type noMethod VoidResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// method id "share.accessfiles":

type AccessfilesCall struct {
	s                 *Service
	fileaccessrequest *FileAccessRequest
	urlParams_        gensupport.URLParams
	ctx_              context.Context
}

// Accessfiles:
func (s *Service) Accessfiles(fileaccessrequest *FileAccessRequest) *AccessfilesCall {
	c := &AccessfilesCall{s: s, urlParams_: make(gensupport.URLParams)}
	c.fileaccessrequest = fileaccessrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *AccessfilesCall) Fields(s ...googleapi.Field) *AccessfilesCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *AccessfilesCall) Context(ctx context.Context) *AccessfilesCall {
	c.ctx_ = ctx
	return c
}

func (c *AccessfilesCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.fileaccessrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "accessfiles")
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

// Do executes the "share.accessfiles" call.
// Exactly one of *FileAccessResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *FileAccessResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *AccessfilesCall) Do() (*FileAccessResponse, error) {
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
	ret := &FileAccessResponse{
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
	//   "id": "share.accessfiles",
	//   "path": "accessfiles",
	//   "request": {
	//     "$ref": "FileAccessRequest",
	//     "parameterName": "resource"
	//   },
	//   "response": {
	//     "$ref": "FileAccessResponse"
	//   }
	// }

}

// method id "share.appendfiles":

type AppendfilesCall struct {
	s                  *Service
	appendfilesrequest *AppendFilesRequest
	urlParams_         gensupport.URLParams
	ctx_               context.Context
}

// Appendfiles:
func (s *Service) Appendfiles(appendfilesrequest *AppendFilesRequest) *AppendfilesCall {
	c := &AppendfilesCall{s: s, urlParams_: make(gensupport.URLParams)}
	c.appendfilesrequest = appendfilesrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *AppendfilesCall) Fields(s ...googleapi.Field) *AppendfilesCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *AppendfilesCall) Context(ctx context.Context) *AppendfilesCall {
	c.ctx_ = ctx
	return c
}

func (c *AppendfilesCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.appendfilesrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "appendfiles")
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

// Do executes the "share.appendfiles" call.
// Exactly one of *UploadFilesResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *UploadFilesResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *AppendfilesCall) Do() (*UploadFilesResponse, error) {
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
	ret := &UploadFilesResponse{
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
	//   "id": "share.appendfiles",
	//   "path": "appendfiles",
	//   "request": {
	//     "$ref": "AppendFilesRequest",
	//     "parameterName": "resource"
	//   },
	//   "response": {
	//     "$ref": "UploadFilesResponse"
	//   }
	// }

}

// method id "share.appendnotes":

type AppendnotesCall struct {
	s                  *Service
	appendnotesrequest *AppendNotesRequest
	urlParams_         gensupport.URLParams
	ctx_               context.Context
}

// Appendnotes:
func (s *Service) Appendnotes(appendnotesrequest *AppendNotesRequest) *AppendnotesCall {
	c := &AppendnotesCall{s: s, urlParams_: make(gensupport.URLParams)}
	c.appendnotesrequest = appendnotesrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *AppendnotesCall) Fields(s ...googleapi.Field) *AppendnotesCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *AppendnotesCall) Context(ctx context.Context) *AppendnotesCall {
	c.ctx_ = ctx
	return c
}

func (c *AppendnotesCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.appendnotesrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "appendnotes")
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

// Do executes the "share.appendnotes" call.
// Exactly one of *VoidResponse or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *VoidResponse.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *AppendnotesCall) Do() (*VoidResponse, error) {
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
	ret := &VoidResponse{
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
	//   "id": "share.appendnotes",
	//   "path": "appendnotes",
	//   "request": {
	//     "$ref": "AppendNotesRequest",
	//     "parameterName": "resource"
	//   },
	//   "response": {
	//     "$ref": "VoidResponse"
	//   }
	// }

}

// method id "share.confirmupload":

type ConfirmuploadCall struct {
	s                    *Service
	confirmuploadrequest *ConfirmUploadRequest
	urlParams_           gensupport.URLParams
	ctx_                 context.Context
}

// Confirmupload:
func (s *Service) Confirmupload(confirmuploadrequest *ConfirmUploadRequest) *ConfirmuploadCall {
	c := &ConfirmuploadCall{s: s, urlParams_: make(gensupport.URLParams)}
	c.confirmuploadrequest = confirmuploadrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ConfirmuploadCall) Fields(s ...googleapi.Field) *ConfirmuploadCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ConfirmuploadCall) Context(ctx context.Context) *ConfirmuploadCall {
	c.ctx_ = ctx
	return c
}

func (c *ConfirmuploadCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.confirmuploadrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "confirmupload")
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

// Do executes the "share.confirmupload" call.
// Exactly one of *VoidResponse or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *VoidResponse.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ConfirmuploadCall) Do() (*VoidResponse, error) {
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
	ret := &VoidResponse{
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
	//   "id": "share.confirmupload",
	//   "path": "confirmupload",
	//   "request": {
	//     "$ref": "ConfirmUploadRequest",
	//     "parameterName": "resource"
	//   },
	//   "response": {
	//     "$ref": "VoidResponse"
	//   }
	// }

}

// method id "share.create":

type CreateCall struct {
	s          *Service
	createreq  *CreateReq
	urlParams_ gensupport.URLParams
	ctx_       context.Context
}

// Create:
func (s *Service) Create(createreq *CreateReq) *CreateCall {
	c := &CreateCall{s: s, urlParams_: make(gensupport.URLParams)}
	c.createreq = createreq
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *CreateCall) Fields(s ...googleapi.Field) *CreateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *CreateCall) Context(ctx context.Context) *CreateCall {
	c.ctx_ = ctx
	return c
}

func (c *CreateCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.createreq)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "create")
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

// Do executes the "share.create" call.
// Exactly one of *CreateResp or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *CreateResp.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *CreateCall) Do() (*CreateResp, error) {
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
	ret := &CreateResp{
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
	//   "id": "share.create",
	//   "path": "create",
	//   "request": {
	//     "$ref": "CreateReq",
	//     "parameterName": "resource"
	//   },
	//   "response": {
	//     "$ref": "CreateResp"
	//   }
	// }

}

// method id "share.delete":

type DeleteCall struct {
	s                     *Service
	deletesnapshotrequest *DeleteSnapshotRequest
	urlParams_            gensupport.URLParams
	ctx_                  context.Context
}

// Delete:
func (s *Service) Delete(deletesnapshotrequest *DeleteSnapshotRequest) *DeleteCall {
	c := &DeleteCall{s: s, urlParams_: make(gensupport.URLParams)}
	c.deletesnapshotrequest = deletesnapshotrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *DeleteCall) Fields(s ...googleapi.Field) *DeleteCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *DeleteCall) Context(ctx context.Context) *DeleteCall {
	c.ctx_ = ctx
	return c
}

func (c *DeleteCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.deletesnapshotrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "delete")
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

// Do executes the "share.delete" call.
// Exactly one of *VoidResponse or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *VoidResponse.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *DeleteCall) Do() (*VoidResponse, error) {
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
	ret := &VoidResponse{
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
	//   "id": "share.delete",
	//   "path": "delete",
	//   "request": {
	//     "$ref": "DeleteSnapshotRequest",
	//     "parameterName": "resource"
	//   },
	//   "response": {
	//     "$ref": "VoidResponse"
	//   }
	// }

}

// method id "share.getmy":

type GetmyCall struct {
	s                *Service
	getsharesrequest *GetSharesRequest
	urlParams_       gensupport.URLParams
	ctx_             context.Context
}

// Getmy:
func (s *Service) Getmy(getsharesrequest *GetSharesRequest) *GetmyCall {
	c := &GetmyCall{s: s, urlParams_: make(gensupport.URLParams)}
	c.getsharesrequest = getsharesrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *GetmyCall) Fields(s ...googleapi.Field) *GetmyCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *GetmyCall) Context(ctx context.Context) *GetmyCall {
	c.ctx_ = ctx
	return c
}

func (c *GetmyCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.getsharesrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "getmy")
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

// Do executes the "share.getmy" call.
// Exactly one of *GetSharesResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *GetSharesResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *GetmyCall) Do() (*GetSharesResponse, error) {
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
	ret := &GetSharesResponse{
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
	//   "id": "share.getmy",
	//   "path": "getmy",
	//   "request": {
	//     "$ref": "GetSharesRequest",
	//     "parameterName": "resource"
	//   },
	//   "response": {
	//     "$ref": "GetSharesResponse"
	//   }
	// }

}

// method id "share.requestupload":

type RequestuploadCall struct {
	s                  *Service
	uploadfilesrequest *UploadFilesRequest
	urlParams_         gensupport.URLParams
	ctx_               context.Context
}

// Requestupload:
func (s *Service) Requestupload(uploadfilesrequest *UploadFilesRequest) *RequestuploadCall {
	c := &RequestuploadCall{s: s, urlParams_: make(gensupport.URLParams)}
	c.uploadfilesrequest = uploadfilesrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *RequestuploadCall) Fields(s ...googleapi.Field) *RequestuploadCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *RequestuploadCall) Context(ctx context.Context) *RequestuploadCall {
	c.ctx_ = ctx
	return c
}

func (c *RequestuploadCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.uploadfilesrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "requestupload")
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

// Do executes the "share.requestupload" call.
// Exactly one of *UploadFilesResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *UploadFilesResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *RequestuploadCall) Do() (*UploadFilesResponse, error) {
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
	ret := &UploadFilesResponse{
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
	//   "id": "share.requestupload",
	//   "path": "requestupload",
	//   "request": {
	//     "$ref": "UploadFilesRequest",
	//     "parameterName": "resource"
	//   },
	//   "response": {
	//     "$ref": "UploadFilesResponse"
	//   }
	// }

}

// method id "share.uploadstatus":

type UploadstatusCall struct {
	s                   *Service
	uploadstatusrequest *UploadStatusRequest
	urlParams_          gensupport.URLParams
	ctx_                context.Context
}

// Uploadstatus:
func (s *Service) Uploadstatus(uploadstatusrequest *UploadStatusRequest) *UploadstatusCall {
	c := &UploadstatusCall{s: s, urlParams_: make(gensupport.URLParams)}
	c.uploadstatusrequest = uploadstatusrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *UploadstatusCall) Fields(s ...googleapi.Field) *UploadstatusCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *UploadstatusCall) Context(ctx context.Context) *UploadstatusCall {
	c.ctx_ = ctx
	return c
}

func (c *UploadstatusCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.uploadstatusrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "uploadstatus")
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

// Do executes the "share.uploadstatus" call.
// Exactly one of *VoidResponse or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *VoidResponse.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *UploadstatusCall) Do() (*VoidResponse, error) {
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
	ret := &VoidResponse{
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
	//   "id": "share.uploadstatus",
	//   "path": "uploadstatus",
	//   "request": {
	//     "$ref": "UploadStatusRequest",
	//     "parameterName": "resource"
	//   },
	//   "response": {
	//     "$ref": "VoidResponse"
	//   }
	// }

}
