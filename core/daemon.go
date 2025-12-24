package core

import (
	"flag"
	"fmt"
	"os"

	"github.com/chuccp/go-web-frame/log"
	"github.com/kardianos/service"
)

type AppService interface {
	Start() error
	Close() error
}

type AppDaemon struct {
	appService AppService
}

func (a *AppDaemon) Start(s service.Service) error {
	go func() {
		err := a.appService.Start()
		if err != nil {
			log.Errors("Failed to start the Daemon service", err)
		}
	}()
	return nil
}

func (a *AppDaemon) Stop(s service.Service) error {
	return a.appService.Close()
}

func RunDaemon(appService AppService, svcConfig *service.Config) {
	// 解析启停参数
	stopFlag := flag.Bool("stop", false, "停止服务")
	flag.Parse()
	app := &AppDaemon{
		appService: appService,
	}
	svc, err := service.New(app, svcConfig)
	if err != nil {
		log.Errors("Failed to create the Daemon service", err)
		os.Exit(1)
	}
	// 处理停止指令
	if *stopFlag {
		if err := svc.Stop(); err != nil {
			log.Errors("Failed to stop the Daemon service", err)
			os.Exit(1)
		}
		fmt.Println("The Daemon service has been stopped")
		os.Exit(0)
	}

	// 启动服务（Windows：注册为服务；Linux：systemd；macOS：launchd）
	if err := svc.Run(); err != nil {
		log.Errors("Failed to run the Daemon service", err)
		os.Exit(1)
	}

}
