package upload

import (
	"appscope.net/api/share/v1"
	"appscope.net/util"

	"github.com/golang/glog"

	"net/http"
	"os"
	"path"
)

type UploadService interface {
	Add(folder string, files []FileInfo, deviceToken, shareToken []byte)
}

type FileInfo struct {
	Name, MimeType string
}

var svc = &service{director: util.NewJobDirector("Upload", 100)}

func GetService() UploadService {
	return svc
}

type service struct {
	director util.JobDirector
}

func (svc *service) Add(folder string, files []FileInfo, deviceToken, shareToken []byte) {
	go util.SafeRun(func() { svc.beginUpload(folder, files, "http", deviceToken, shareToken) })
}

func (uploadService *service) beginUpload(folder string,
	files []FileInfo, tag string,
	deviceToken, shareToken []byte) {
	// first allocate everything
	httpClient := &http.Client{}
	svc, err := share.New(httpClient)
	if err != nil {
		glog.Error(err)
		return
	}

	fileReq := &share.UploadFilesRequest{
		DeviceToken: deviceToken,
		ShareToken:  shareToken,
		Files:       make([]*share.FileItem, 0, len(files))}

	for _, file := range files {
		if fi, err := os.Stat(path.Join(folder, file.Name)); err != nil {
			glog.Errorf("Couldn't stat %s: %v", path.Join(folder, file.Name), err)
			return
		} else {
			fileReq.Files = append(fileReq.Files, &share.FileItem{
				Name:     file.Name,
				MimeType: file.MimeType,
				Size:     fi.Size(),
				Tag:      tag,
			})
		}
	}

	resp, err := svc.Requestupload(fileReq).Do()
	if err != nil {
		glog.Error(err)
		return
	} else if resp.Error != nil {
		glog.Error(resp.Error)
		return
	}

	errch := make(chan error)
	nfiles := len(files)
	go util.SafeRun(func() {
		for i := 1; i <= nfiles; i++ {
			<-errch
		}
		// remove folder
		glog.Infof("Upload of %s complete", folder)
	})

	for _, file := range resp.Files {
		uploadService.director.Add(&GCSUpload{
			client:   httpClient,
			file:     file,
			filePath: path.Join(folder, file.Name)},
			GoogleCloudStorageUpload,
			errch)
	}

}
