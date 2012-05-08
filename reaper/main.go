package main

import (
	"flag"
	"log"
	"net/rpc"
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
	err = c.Call("Server.Start", username, nil)
	if err != nil {
		return err
	}
	if !ok {
		log.Println("Start failed")
		shutdown()
		return nil
	}
	for err == nil {
		time.Sleep(checkinInterval)
		err = c.Call("Server.CheckIn", username, &ok)
		if err == nil && !ok {
			log.Println("CheckIn failed.")
			shutdown()
			return nil
		}
	}
	log.Println(err)
	return c.Call("Server.Stop", username, nil)
}

func shutdown() {
	log.Println("shutting down machine")
	// TODO(adg): actually shut down
}
