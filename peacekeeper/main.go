package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"time"
)

var (
	dryRun     = flag.Bool("dry", false, "don't actually shutdown/sleep")
	interval   = flag.Duration("interval", time.Second*10, "checkin interval")
	serverAddr = flag.String("server", "localhost:7020", "server address")
	sleepOnly  = flag.Bool("sleep", false, "sleep instead of shutting down")
	username   = flag.String("user", "", "user name")
)

var (
	transport = &http.Transport{MaxIdleConnsPerHost: 1}
	client    = &http.Client{Transport: transport}
)

func main() {
	flag.Parse()
	for {
		if err := checkin(); err != nil {
			log.Println("checkin:", err)
		}
		time.Sleep(*interval)
	}
}

func checkin() error {
	u := fmt.Sprintf("http://%s/checkin", *serverAddr)
	v := url.Values{"username": []string{*username}}
	r, err := client.PostForm(u, v)
	if err != nil {
		return err
	}
	r.Body.Close()
	if r.StatusCode == http.StatusForbidden {
		transport.CloseIdleConnections()
		shutdown()
		time.Sleep(*interval) // sleep twice after shutdown
	}
	return nil
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
