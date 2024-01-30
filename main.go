package main

import (
	"challenge/client"
	"challenge/server"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	go server.Start()
	fmt.Println("Running until user hits ctrl+c...")
	go func() {
		for range time.Tick(time.Minute) {
			client.Start()
		}
	}()
	<-done
}
