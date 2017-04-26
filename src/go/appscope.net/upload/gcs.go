package upload

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"appscope.net/api/share/v1"
	"github.com/golang/glog"
)

/*
 * google cloud storage upload
 * requires an up to date file
 */
type GCSUpload struct {
	client   *http.Client
	filePath string

	file *share.UploadFileItem
}

/*
 * https://cloud.google.com/storage/docs/json_api/v1/how-tos/upload#upload-resumable
 *
 * 1. try to upload file at once
 * 2. retry in case of error by sending the remaining bytes
 * 3. notify backend on completion
 */
func GoogleCloudStorageUpload(v interface{}) (err error) {
	gu := v.(*GCSUpload)
	finfo, err := os.Stat(gu.filePath)
	if err != nil {
		return err
	}

	fileSize := finfo.Size()

	if fileSize != gu.file.Size {
		return fmt.Errorf("Actual (%d) file %s size doesn't match declared (%d)", fileSize, gu.filePath, gu.file.Size)
	}

	file, err := os.Open(gu.filePath)
	if err != nil { // unrecoverable error
		return err
	}
	defer file.Close()

	var offset int64 = 0

	if off, err := file.Seek(offset, 0); err != nil || off != offset {
		return fmt.Errorf("seek failed : %v %v", err, off)
	}

	req, err := http.NewRequest("PUT", gu.file.Url, file)
	if err != nil {
		return fmt.Errorf("PUT %s : %v", gu.file.Url, err)
	}

	if gu.file.MimeType != "" {
		req.Header.Set("Content-Type", gu.file.MimeType)
	} else {
		req.Header.Set("Content-Type", "application/octet-stream")
	}

	//req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d",
	//	offset+1, fileSize-offset, fileSize))
	//req.ContentLength = fileSize - offsets

	tUploadStart := time.Now()
	res, err := gu.client.Do(req)
	if err != nil {
		glog.Errorf("Failed upload %s to %s : %v", gu.filePath, gu.file.Url, err)
	} else if glog.V(vUploadProgress) {
		glog.Infof("Uploaded %s, %d bytes in %s", gu.filePath, fileSize, time.Since(tUploadStart).String())

	}

	if glog.V(vUploadRequestTroubleshooting) {
		dd, _ := httputil.DumpRequest(req, false)
		glog.Infof("%s : REQUEST %s", gu.filePath, string(dd))
		dd, _ = httputil.DumpResponse(res, true)
		glog.Infof("%s : RESPONSE %s", gu.filePath, string(dd))
	}

	return err
}

func (gu *GCSUpload) requestCurrentRange() (offset int64, err error, recoverable bool) {
	return 0, nil, true
}

func (gu *GCSUpload) ToString() string {
	return fmt.Sprintf("%s -> %s", gu.filePath, gu.file.Url)
}
