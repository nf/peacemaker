package main

import (
	"flag"
	"log"
	"net/http"
	"net/rpc"
)

var (
	datafile = flag.String("data", "user.json", "user data file")
	httpAddr = flag.String("http", ":7020", "HTTP server listen address")
)

var server *Server

func main() {
	flag.Parse()
	var err error
	server, err = NewServer(*datafile)
	if err != nil {
		log.Fatal(err)
	}
	rpc.Register(server)
	rpc.HandleHTTP()
	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}
