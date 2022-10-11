package main

import (
	"apigateway/config"
	"apigateway/discovery"
	"apigateway/internal/service"
	"apigateway/routes"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

func main() {
	config.InitConfig()
	// 服务发现
	etcdAdress := []string{viper.GetString("etcd.address")}
	etcdRegister := discovery.NewResolver(etcdAdress, logrus.New())
	resolver.Register(etcdRegister)
	go startListen()
	{
		osSignal := make(chan os.Signal, 1)
		signal.Notify(osSignal, os.Interrupt, os.Kill, syscall.SIGTERM)
		<-osSignal
		fmt.Println("exit")
	}
}

func startListen() {
	// grpc client
	opts := []grpc.DialOption{}
	userConn, _ := grpc.Dial("127.0.0.1:10001", opts...)
	userService := service.NewUserServiceClient(userConn)

	ginRouter := routes.NewRouter(userService)
	server := &http.Server{
		Addr:              viper.GetString("service.port"),
		Handler:           ginRouter,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("绑定失败", err)
	}
}
