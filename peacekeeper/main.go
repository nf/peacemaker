package main

import (
	"flag"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"time"
)

const checkinInterval = time.Second * 10

var (
	serverAddr = flag.String("server", "localhost:7020", "server address")
	username   = flag.String("user", "", "user name")
	dryRun     = flag.Bool("dry", false, "don't actually shutdown/sleep")
	sleepOnly  = flag.Bool("sleep", false, "sleep instead of shutting down")
)

var signals = make(chan os.Signal)

func main() {
	flag.Parse()
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

	for err == nil {
		var ok bool
		err = c.Call("Server.CheckIn", username, &ok)
		if err == nil {
			if !ok {
				shutdown()
				return nil
			}
			time.Sleep(checkinInterval)
		}
	}
	return err
}

func shutdown() {
	log.Println("shutting down machine")
	if *dryRun {
		log.Println("(not actually shutting down)")
		return
	}
	opt := "-h"
	if *sleepOnly {
		opt = "-s"
	}
	err := exec.Command("shutdown", opt, "now").Run()
	if err != nil {
		log.Println("shutdown:", err)
	}
}
