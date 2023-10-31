package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	openapi "test-flag/openxpanapi"
)

func main() {
	req := LoginCode{}
	input := bufio.NewReader(os.Stdin)
	var err error
	fmt.Println("请输入 code")
	req.Code, err = input.ReadString('\n')
	req.Code = req.Code[:len(req.Code)-1]
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
