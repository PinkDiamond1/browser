package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/browser/db"
	"github.com/browser/dispatch"
	_ "github.com/browser/init"
)

func main() {
	SignalHandler()
	// go statis.Start()
	dispatch.NewDispatch().Start()
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
