package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/kardianos/service"
)

type AppService struct{}

func (a *AppService) Start(s service.Service) error {
	go a.run()
	return nil
}

func (a *AppService) Stop(s service.Service) error {
	fmt.Println("优雅关闭中...")
	time.Sleep(1 * time.Second)
	return nil
}

func (a *AppService) run() {
	// 业务逻辑
	for {
		fmt.Println("应用运行中...", time.Now())
		time.Sleep(2 * time.Second)
	}
}

func main() {
	// 解析启停参数
	stopFlag := flag.Bool("stop", false, "停止服务")
	flag.Parse()

	// 配置服务
	svcConfig := &service.Config{
		Name:        "MyApp",
		DisplayName: "My Application",
		Description: "My Go Application with Service Management",
	}

	app := &AppService{}
	svc, err := service.New(app, svcConfig)
	if err != nil {
		fmt.Printf("创建服务失败：%v\n", err)
		os.Exit(1)
	}

	// 日志适配
	logger, err := svc.Logger(nil)
	if err != nil {
		fmt.Printf("日志初始化失败：%v\n", err)
		os.Exit(1)
	}

	// 处理停止指令
	if *stopFlag {
		if err := svc.Stop(); err != nil {
			logger.Error("停止服务失败：", err)
			os.Exit(1)
		}
		fmt.Println("服务已停止")
		os.Exit(0)
	}

	// 启动服务（Windows：注册为服务；Linux：systemd；macOS：launchd）
	if err := svc.Run(); err != nil {
		logger.Error("服务运行失败：", err)
		os.Exit(1)
	}

}
