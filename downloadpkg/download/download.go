package download

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"test-flag/downloadpkg/utils"
	"time"
	// "icode.baidu.com/baidu/xpan/go-sdk/xpan/utils"
)

const KB = 1024
const MB = 1024 * KB

var queueChannel chan struct{}

func Download(accessToken string, dlink string, outputFilename string, size uint64) error {
	uri := dlink + "&" + "access_token=" + accessToken
	begin := time.Now()
	queueChannel = make(chan struct{}, runtime.NumCPU())
	// 创建一个通道用于接收信号
	var SigCh chan os.Signal
	SigCh = make(chan os.Signal, 1)
	ctx, cancelFunc := context.WithCancel(context.Background())
	// 捕获 SIGINT（Ctrl+C） 和 SIGTERM（kill 命令）
	signal.Notify(SigCh, syscall.SIGINT)
	go func() {
		select {
		case <-SigCh:
			logrus.Info("下载终止")
			cancelFunc()
			return
		}
	}()
	switch {
	case size > 100*MB:

		sum := size / (100 * MB)
		var wg sync.WaitGroup
		logrus.Info("共临时文件", sum)
		for i := 0; uint64(i) <= sum; i++ {

			select {
			case <-ctx.Done():
				return errors.New("下载终止")
			default:
				wg.Add(1)
				if uint64(i) == sum {
					go doRequest(uri, uint64(i), 0, outputFilename, true, &wg)
				} else {
					go doRequest(uri, uint64(i), 0, outputFilename, false, &wg)
				}
			}
		}
		wg.Wait()
		logrus.Info("等待结束")
		file, err := os.OpenFile(outputFilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			fmt.Printf("无法写入文件 %s: %v\n", outputFilename, err)
			return err
		}
		for i := 0; uint64(i) <= sum; i++ {
			filename := outputFilename + strconv.FormatUint(uint64(i), 10)
			content, err := ioutil.ReadFile(filename)
			if err != nil {
				fmt.Printf("无法读取文件 %s: %v\n", filename, err)
				return err
			}
			logrus.Info("开始拼接")
			_, err = file.Write(content)
			if err != nil {
				return err
			}
			go func() {
				err := os.Remove(filename)
				if err != nil {
				}
			}()
			if err != nil {
				return err
			}
		}
	default:
		headers := map[string]string{
			"User-Agent": "pan.baidu.com",
		}

		var postBody io.Reader
		body, statusCode, err := utils.Do2HTTPRequest(uri, postBody, headers)
		if err != nil {
			return err
		}
		if statusCode != 200 {
			return errors.New("download http fail")
		}

		// 下载数据输出到名“outputFilename”的文件
		file, err := os.OpenFile(outputFilename, os.O_WRONLY|os.O_CREATE, 0666)
		defer file.Close()
		write := bufio.NewWriter(file)
		_, err = write.WriteString(body)
		if err != nil {
			return err
		}
		//Flush将缓存的文件真正写入到文件中
		err = write.Flush()
		if err != nil {
			return err
		}

		return nil
	}

	logrus.Info("下载共计使用时间:", time.Since(begin))
	return nil
}

func doRequest(uri string, index uint64, restart int, filename string, isEnd bool, wg *sync.WaitGroup) {
	fileInfo, err := os.Stat(filename + strconv.FormatUint(index, 10))

	if err == nil && fileInfo.Size() == int64(100*MB) {
		logrus.Info("切片文件:", filename+strconv.FormatUint(index, 10), "已存在且完整，跳过下载此切片文件")
		wg.Done()
		return
	}
	queueChannel <- struct{}{}
	time.Sleep(time.Duration(len(queueChannel)) * 1500 * time.Millisecond)
	logrus.Info("开始携程工作：", index)
	headers := map[string]string{
		"User-Agent": "pan.baidu.com",
	}
	if isEnd {
		headers["Range"] = "bytes=" + strconv.FormatUint(100*MB*index, 10) + "-"
	} else {
		headers["Range"] = "bytes=" + strconv.FormatUint(100*MB*index, 10) + "-" + strconv.FormatUint(100*MB*(index+1)-1, 10)
	}
	var postBody io.Reader
	body, statusCode, err := utils.Do2HTTPRequest(uri, postBody, headers)
	logrus.Error(statusCode, err)
	if err != nil {
		logrus.Error(err)
		logrus.Info("开始重新下载文件,下载编号: ", index, " 重载次数: ", restart)
		if restart < 3 {
			time.Sleep(2 * time.Duration(restart) * time.Second)
		} else {
			time.Sleep(2 * time.Duration(restart) * time.Second)
		}
		go doRequest(uri, index, restart+1, filename, isEnd, wg)
		return
	}

	if statusCode != 200 && statusCode != 206 {
		logrus.Error(err)

		logrus.Info("开始重新下载文件,下载编号: ", index, " 重载次数: ", restart)
		if restart < 3 {
			time.Sleep(2 * time.Duration(restart) * time.Second)
		} else {
			time.Sleep(2 * time.Duration(restart) * time.Second)
		}
		go doRequest(uri, index, restart+1, filename, isEnd, wg)
		return
	}
	// 下载数据输出到名“outputFilename”的文件
	file, err := os.OpenFile(filename+strconv.FormatUint(index, 10), os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	write := bufio.NewWriter(file)
	_, err = write.WriteString(body)
	if err != nil {
		logrus.Error(err)
		logrus.Info("开始重新下载文件,下载编号: ", index, " 重载次数: ", restart)
		if restart < 3 {
			time.Sleep(2 * time.Duration(restart) * time.Second)
		} else {
			time.Sleep(2 * time.Duration(restart) * time.Second)
		}
		go doRequest(uri, index, restart+1, filename, isEnd, wg)
		return
	}
	//Flush将缓存的文件真正写入到文件中
	err = write.Flush()
	if err != nil {
		logrus.Error(err)
		logrus.Info("开始重新下载文件,下载编号: ", index, " 重载次数: ", restart)
		if restart < 3 {
			time.Sleep(2 * time.Duration(restart) * time.Second)
		} else {
			time.Sleep(2 * time.Duration(restart) * time.Second)
		}
		go doRequest(uri, index, restart+1, filename, isEnd, wg)
		return
	}
	<-queueChannel
	wg.Done()
}
