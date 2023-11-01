package main

import (
	"context"
	"flag"
	"github.com/sirupsen/logrus"
	openapi "test-flag/openxpanapi"
	"unicode/utf8"
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
	response, _, err := apiClient.FileinfoApi.Xpanfilelist(context.Background()).AccessToken(req.AccessToken).Dir(req.Path).Execute()
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
	var err error
	flag.StringVar(&req.AccessToken, "access_token", "", "设置access_token")
	flag.StringVar(&req.Path, "path", "", "设置文件路径")
	flag.Parse()
	logrus.Infof("%+v", req)
	info, err := req.ShowDirInfo()
	if err != nil {
		logrus.Error(err)
		return
	}

	logrus.Info(decodeUnicode(*info))
}

func decodeUnicode(s string) string {
	rs := []rune(s)
	for i := 0; i < len(rs); {
		r := rs[i]
		if r == '\\' && i < len(rs)-1 && rs[i+1] == 'u' {
			r, size := utf8.DecodeRuneInString(s[i+2:])
			rs = append(rs[:i], append([]rune(string(r)), rs[i+size+2:]...)...)
			i += size
		} else {
			i++
		}
	}
	return string(rs)
}
