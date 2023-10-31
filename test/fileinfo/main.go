package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"net/url"
	openapi "test-flag/openxpanapi"
)

type ShowDirInfoReq struct {
	Path        string `json:"path"`
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
func (req ShowDirInfoReq) ShowDirInfo() (*string, error) {
	response, _, err := apiClient.FileinfoApi.Xpanfilelist(context.Background()).AccessToken(req.AccessToken).Dir(url.PathEscape(req.Path)).Execute()
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func main() {
	req := ShowDirInfoReq{
		Path:        "",
		AccessToken: "",
	}
	info, err := req.ShowDirInfo()
	if err != nil {
		logrus.Error(err)
		return
	}
	logrus.Info(info)
}
