package main

import (
	"bufio"
		"net/url"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
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
	input := bufio.NewReader(os.Stdin)
	var err error
	fmt.Println("请输入 access_token")
	req.AccessToken, err = input.ReadString('\n')
	fmt.Println("请输入 path 路径")
	req.Path, err = input.ReadString('\n')
	fmt.Println("请输入 fs_id")
	str, err := input.ReadString('\n')
	req.FsID, err = strconv.ParseUint(str, 10, 64)
	fmt.Println("请输入 is_dir")
	str, err = input.ReadString('\n')
	req.IsDir, err = strconv.ParseBool(str)
	err = req.Download()
	if err != nil {
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
		response, _, err := apiClient.MultimediafileApi.Xpanfilelistall(context.Background()).Recursion(1).Path(url.PathEscape(req.Path)).Execute()
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
		for _, meta := range metas.List {
			err := download.Download(req.AccessToken, meta.Dlink, meta.Filename, meta.Size)
			if err != nil {
				logrus.Error(err)
				return err
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
			err := download.Download(req.AccessToken, meta.Dlink, meta.Filename, meta.Size)
			if err != nil {
				logrus.Error(err)
				return err
			}
		}
	}
	return nil
}
