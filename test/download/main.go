package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/sirupsen/logrus"
	"os"
	"test-flag/downloadpkg/download"
	openapi "test-flag/openxpanapi"
)

func main() {
	req := DownloadInfoReq{
		IsDir:       false,
		Path:        "",
		FsID:        0,
		AccessToken: "",
	}
	var err error
	flag.StringVar(&req.AccessToken, "access_token", "", "登录凭证")
	flag.StringVar(&req.Path, "path", "", "路径")
	flag.Uint64Var(&req.FsID, "fs_id", 0, "fs_id")

	flag.BoolVar(&req.IsDir, "is_dir", false, "是否是文件夹")
	flag.Parse()
	logrus.Infof("%+v", req)
	err = req.Download()
	if err != nil {
		logrus.Error(err)
		return
	}
}

type DownloadInfoReq struct {
	IsDir       bool   `json:"is_dir"`
	Path        string `json:"path"`
	FsID        uint64 `json:"fs_id,string"`
	AccessToken string `json:"access_token"`
}

var (
	configuration *openapi.Configuration
	apiClient     *openapi.APIClient
)

func init() {
	configuration = openapi.NewConfiguration()
	apiClient = openapi.NewAPIClient(configuration)
}

func (req DownloadInfoReq) Download() error {
	switch req.IsDir {
	case true:
		response, _, err := apiClient.MultimediafileApi.Xpanfilelistall(context.Background()).AccessToken(req.AccessToken).Recursion(1).Path(req.Path).Execute()
		if err != nil {
			logrus.Error(err)
			return err
		}
		var readFileListRespBody struct {
			Errno int `json:"errno"`
			List  []struct {
				FileName string `json:"server_filename"`
				FsID     int64  `json:"fs_id"`
			} `json:"list"`
		}
		err = json.Unmarshal([]byte(response), &readFileListRespBody)
		if err != nil {
			logrus.Error(err)
			return err
		}
		var fsIDs = make([]uint64, 0)
		for _, fileInfo := range readFileListRespBody.List {
			fsIDs = append(fsIDs, uint64(fileInfo.FsID))
		}
		metasArg := download.NewFileMetasArg(fsIDs, "./")
		metas, err := download.FileMetas(req.AccessToken, metasArg)
		if err != nil {
			logrus.Error(err)
			return err
		}
		logrus.Info("file num: ", len(metas.List))
		for _, meta := range metas.List {
			logrus.Info("dlink:", meta.Dlink, " filename:", meta.Filename, " size:", meta.Size)
		}
		logrus.Info("開始下載")
		for _, meta := range metas.List {
			if meta.Isdir == 0 {
				logrus.Info(meta.Path)
				logrus.Info(meta.Filename)
				logrus.Info(meta.Path[:len(meta.Path)-len(meta.Filename)])
				path := "." + meta.Path[len(req.Path):len(meta.Path)-len(meta.Filename)]
				logrus.Info(path)
				err := os.MkdirAll(path, 0777)
				if err != nil {
					return err
				}
				err = download.Download(path, req.AccessToken, meta.Dlink, meta.Filename, meta.Size)
				if err != nil {
					logrus.Error(err)
					return err
				}
			}
		}
	case false:
		fsIDs := []uint64{req.FsID}
		metasArg := download.NewFileMetasArg(fsIDs, "./")
		metas, err := download.FileMetas(req.AccessToken, metasArg)
		if err != nil {
			logrus.Error(err)
			return err
		}
		for _, meta := range metas.List {
			logrus.Info(meta.Size)
			logrus.Info(meta.Path)
			err := download.Download("", req.AccessToken, meta.Dlink, meta.Filename, meta.Size)

			if err != nil {
				logrus.Error(err)
				return err
			}
		}
	}
	return nil
}
