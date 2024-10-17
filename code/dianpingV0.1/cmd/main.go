package main

import (
	"context"
	"dianping/internal/config"
	"dianping/internal/router"
	"dianping/pkg/logger"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"
)

func init() {
	configPath := pflag.StringP("config", "c", "../configs/config.yaml", "config file path")
	pflag.Parse()

	config.InitConfig(*configPath) //初始化配置

	//初始化日志
	err := logger.InitLogger(config.LogOption)
	if err != nil {
		panic(err)
	}
	//初始化数据库
	// db.DBEngine, err = db.NewMySQL(config.MysqlOption)
	// if err != nil {
	// 	panic(err)
	// }

}

func main() {
	r := router.NewRouter()

	// err := r.Run(":" + config.ServerOption.HttpPort)
	// if err != nil {
	// 	panic(err)
	// }

	//创建HTTP服务器
	server := http.Server{
		Addr:    ":" + config.ServerOption.HttpPort,
		Handler: r,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // syscall.SIGKILL是无法捕捉的
	<-quit
	fmt.Println("shutdown server...")

	//创建超时上下文，Shutdown可以让未处理的连接在这个时间内关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		panic(err)
	}
	fmt.Println("server shutdown success")
}
