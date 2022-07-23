package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/binjip978/tftp/server"
)

func main() {
	filePath := flag.String("file", "", "file to serve via tftp (required)")
	addr := flag.String("addr", "0.0.0.0", "server address")
	port := flag.Int("port", 6900, "tftp server port")
	flag.Parse()

	_, err := os.Stat(*filePath)
	if err != nil {
		log.Fatal(err)
	}

	srv := server.Server{
		Handler: server.FileHandler(*filePath),
		Timeout: 2 * time.Second,
		Retry:   5,
	}

	log.Fatal(srv.ListenAndServe(net.JoinHostPort(*addr, strconv.Itoa(*port))))
}
