package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/binjip978/tftp/server"
)

func main() {
	filePath := flag.String("file", "", "file to serve via tftp (required)")
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

	log.Fatal(srv.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", *port)))
}
