package core

import (
	"flag"
	"fmt"
	"os"

	"github.com/chuccp/go-web-frame/log"
	"github.com/kardianos/service"
)

type AppDaemon struct {
	webFrame *WebFrame
}

func (a *AppDaemon) Start(s service.Service) error {
	go func() {
		err := a.webFrame.Start()
		if err != nil {
			log.Errors("启动服务失败：", err)
		}
	}()
	return nil
}

func (a *AppDaemon) Stop(s service.Service) error {
	return a.webFrame.Close()
}

func Run(webFrame *WebFrame, svcConfig *service.Config) {
	// 解析启停参数
	stopFlag := flag.Bool("stop", false, "停止服务")
	flag.Parse()
	app := &AppDaemon{
		webFrame: webFrame,
	}
	svc, err := service.New(app, svcConfig)
	if err != nil {
		log.Errors("创建服务失败：", err)
		os.Exit(1)
	}
	// 处理停止指令
	if *stopFlag {
		if err := svc.Stop(); err != nil {
			log.Errors("停止服务失败：", err)
			os.Exit(1)
		}
		fmt.Println("服务已停止")
		os.Exit(0)
	}

	// 启动服务（Windows：注册为服务；Linux：systemd；macOS：launchd）
	if err := svc.Run(); err != nil {
		log.Errors("服务运行失败：", err)
		os.Exit(1)
	}

}
