package main

import (
	"flag"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

const checkinInterval = time.Second * 10

var (
	serverAddr = flag.String("server", "localhost:7020", "server address")
	username   = flag.String("user", "", "user name")
	sleepOnly  = flag.Bool("sleep", false, "don't halt; only sleep")
)

var signals = make(chan os.Signal)

func main() {
	flag.Parse()
	signal.Notify(signals, syscall.SIGTERM)
	for {
		if err := run(); err != nil {
			log.Println("run:", err)
		}
		time.Sleep(checkinInterval)
	}
}

func run() error {
	c, err := rpc.DialHTTP("tcp", *serverAddr)
	if err != nil {
		return err
	}
	defer c.Close()

	var ok bool
	err = c.Call("Server.Start", username, &ok)
	if err != nil {
		return err
	}
	if !ok {
		log.Println("Start failed")
		shutdown()
		return nil
	}

	done := make(chan bool)
	go handleSignals(c, done)
	defer func() { done <- true }()

	for err == nil {
		time.Sleep(checkinInterval)
		err = c.Call("Server.CheckIn", username, &ok)
		if err == nil && !ok {
			log.Println("CheckIn failed")
			shutdown()
			return nil
		}
	}
	return err
}

func handleSignals(c *rpc.Client, done chan bool) {
	select {
	case <-signals:
		c.Call("Server.Stop", username, nil)
		c.Close()
		os.Exit(1)
	case <-done:
	}
}

func shutdown() {
	log.Println("shutting down machine")
	opt := "-h"
	if *sleepOnly {
		opt = "-s"
	}
	err := exec.Command("shutdown", opt, "now").Run()
	if err != nil {
		log.Println("shutdown:", err)
	}
}
