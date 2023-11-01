package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	openapi "test-flag/openxpanapi"
)

func main() {
	req := LoginCode{}

	var err error
	// fmt.Println("授权码模式url http://openapi.baidu.com/oauth/2.0/authorize?response_type=code&client_id=zNBhtXeLhZDRoxMI6trDohpVREC5AEFP&redirect_uri=oob&scope=basic,netdisk&device_id=39856593")
	flag.StringVar(&req.Code, "code", "", "授权码")
	flag.Parse()
	logrus.Infof("%+v", req.Code)
	accessToken, err := req.VerifyCode()
	if err != nil {
		logrus.Error(err)
		return
	}
	fmt.Println(accessToken)
}

// LoginCode 百度用户通过授权码进行登录
type LoginCode struct {
	Code string `json:"code"`
}

var (
	configuration *openapi.Configuration
	apiClient     *openapi.APIClient
)

func init() {
	configuration = openapi.NewConfiguration()
	apiClient = openapi.NewAPIClient(configuration)
}
func (code LoginCode) VerifyCode() (string, error) {
	ctx := context.Background()

	resp, _, err := apiClient.AuthApi.OauthTokenCode2token(ctx).Code(code.Code).ClientId("zNBhtXeLhZDRoxMI6trDohpVREC5AEFP").ClientSecret("ZllR6fnf7T7r9qtFpismGmmQ4k4SZ3Ao").RedirectUri("oob").Execute()
	if err != nil {
		return "", err
	}
	return resp.GetAccessToken(), nil

}
