package main

import (
	"flag"
	"log"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const checkinInterval = time.Second * 10

var (
	serverAddr = flag.String("server", "localhost:7020", "server address")
	username   = flag.String("user", "", "user name")
)

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatal(err)
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

	go handleSignals(c)

	for err == nil {
		time.Sleep(checkinInterval)
		err = c.Call("Server.CheckIn", username, &ok)
		if err == nil && !ok {
			log.Println("CheckIn failed")
			shutdown()
			return nil
		}
	}
	log.Println(err)
	return c.Call("Server.Stop", username, nil)
}

func handleSignals(c *rpc.Client) {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM)
	<-ch
	c.Call("Server.Stop", username, nil)
	c.Close()
	os.Exit(1)
}

func shutdown() {
	log.Println("shutting down machine")
	// TODO(adg): actually shut down
}
