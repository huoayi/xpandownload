package config

import (
	"context"
	"os"
)

var (
	Ctx        context.Context
	CannelFunc context.CancelFunc
	SigCh      chan os.Signal
)

func init() {
	Ctx, CannelFunc = context.WithCancel(context.Background())
	// 创建一个通道用于接收信号
	SigCh = make(chan os.Signal, 1)

	// 捕获 SIGINT（Ctrl+C） 和 SIGTERM（kill 命令）
	//	signal.Notify(SigCh, syscall.SIGINT, syscall.SIGTERM)

}
