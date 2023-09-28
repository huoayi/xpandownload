package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"test-flag/config"
	"test-flag/downloadpkg/download"
	openapi "test-flag/openxpanapi"
)

const (
	DOWNLOAD = "download"
	CD       = "cd"
	LS       = "ls"
)

var ApiClient *openapi.APIClient
var AccessToken string

func main() {
	// 简化模式 请访问以下url并截取返回路径中的access_token
	// url http://openapi.baidu.com/oauth/2.0/authorize?response_type=token&client_id=62ezec4ZtP1UbL0ZReOOsv4cWGS6FKLy&redirect_uri=oob&scope=basic,netdisk

	// 授权码模式
	//url http://openapi.baidu.com/oauth/2.0/authorize?response_type=code&client_id=62ezec4ZtP1UbL0ZReOOsv4cWGS6FKLy&redirect_uri=oob&scope=basic,netdisk&device_id=38554407
	// 基于简化模式的获取access_token方式

	fmt.Println("可通过访问该url获取路径参数中的access_token")
	fmt.Println("http://openapi.baidu.com/oauth/2.0/authorize?response_type=token&client_id=62ezec4ZtP1UbL0ZReOOsv4cWGS6FKLy&redirect_uri=oob&scope=basic,netdisk")
	fmt.Println("请输入您获取到的access_token:")
	// TODO: 填充最新的 access_token
	input := bufio.NewReader(os.Stdin)
	var err error
	AccessToken, err = input.ReadString('\n')
	if err != nil {
		logrus.Error(err)
		return
	}
	AccessToken = AccessToken[:len(AccessToken)-1]
	configuration := openapi.NewConfiguration()

	ApiClient = openapi.NewAPIClient(configuration)

	var dir string
	dir = "/"

	var fileDLink map[string]string
	var fileSize map[string]uint64
	fileDLink = make(map[string]string)
	fileSize = make(map[string]uint64)
	fmt.Println("可选择命令：ls, cd, download")
	for {
		fmt.Print(dir, "  :")

		commandLine, err := input.ReadString('\n')
		if err != nil {
			return
		}
		if err != nil {
			logrus.Error(err)
			fmt.Println("输入有误，请重新输入")
			continue
		}
		commandArgs := strings.Split(commandLine, " ")

		switch commandArgs[0] {
		case CD:
			if len(commandArgs) != 2 {
				fmt.Println("输入有误，请重新输入")
				break
			}
			switch commandArgs[1] {
			case "..\n":
				dirs := strings.Split(dir, "/")
				if len(dirs) == 1 {
					fmt.Println("无上一级")
					break
				}
				dir = strings.Join(dirs[:len(dirs)-1], "/")
			case ".\n":

			default:
				if err = GetFileInfo(dir, fileDLink, fileSize); err != nil {
					fmt.Println(err)
					break
				}
				dir += commandArgs[1][:len(commandArgs[1])-1] + "/"
			}

		case LS + "\n":
			err := GetFileInfo(dir, fileDLink, fileSize)
			if err != nil {
				logrus.Error(err)
				break
			}
			break
		case DOWNLOAD:
			if len(commandArgs) != 2 {
				fmt.Println("输入有误，请重新输入")
				break
			}
			ok, err := Download(commandArgs[1][:len(commandArgs[1])-1], fileDLink, fileSize)
			if err != nil || !ok {

				logrus.Error(ok, " ", err)
				break
			}
			fmt.Println("下载完成")
		default:
			fmt.Println("未知命令")
			break
		}
	}

}
func GetFileInfo(dir string, fileDLink map[string]string, fileSize map[string]uint64) error {
	// 按照自己的需求获取某文件夹下所有的文件（可进行排序）
	readFileListResp, _, err := ApiClient.FileinfoApi.Xpanfilelist(config.Ctx).AccessToken(AccessToken).Folder("0").Start("0").Dir(dir).Execute()
	if err != nil {
		logrus.Error(err)
	}

	var readFileListRespBody struct {
		Errno int `json:"errno"`
		List  []struct {
			FileName string `json:"server_filename"`
			FsID     int64  `json:"fs_id"`
		} `json:"list"`
	}
	err = json.Unmarshal([]byte(readFileListResp), &readFileListRespBody)
	if err != nil {
		logrus.Error(err)
		return err
	}
	if readFileListRespBody.Errno != 0 {
		return errors.New("访问错误")
	}
	var ufsid []uint64
	for _, fileBody := range readFileListRespBody.List {
		ufsid = append(ufsid, uint64(fileBody.FsID))
	}
	args := download.NewFileMetasArg(ufsid, "./")
	metas, err := download.FileMetas(AccessToken, args)
	if err != nil {
		return err
	}
	for key := range fileDLink {
		delete(fileDLink, key)
	}

	for _, meta := range metas.List {
		fmt.Println(meta.Filename)
		fileDLink[meta.Filename] = meta.Dlink
		fileSize[meta.Filename] = meta.Size
	}
	return nil
}
func Download(filename string, fileDLink map[string]string, fileSize map[string]uint64) (bool, error) {

	err := download.Download(AccessToken, fileDLink[filename], filename, fileSize[filename])
	if err != nil {
		return false, err
	}
	return true, nil
}
