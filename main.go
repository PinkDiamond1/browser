package main

import (
	"fmt"
	"github.com/browser_service/db"
	"github.com/browser_service/dispatch"
	_ "github.com/browser_service/init"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	dispatch.NewDispatch().Start()
	SignalHandler()
}

func SignalHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c //阻塞等待
		fmt.Println("system exit")
		_ = db.Mysql.Close
		os.Exit(0)
	}()
}
