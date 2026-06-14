package main

import (
	"log"
	"os"
	"os/signal"
	"sync"

	"syscall"

	"github.com/Ozone317/Kova/server"
)

func main() {
	var sigs chan os.Signal = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		if err := server.RunAsyncTCPServer(&wg); err != nil {
			log.Printf("TCP server failed: %v", err)
		}
	}()
	go server.WaitForSignal(&wg, sigs)

	wg.Wait()
}
