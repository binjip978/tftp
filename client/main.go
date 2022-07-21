package main

import (
	"flag"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// cfg defines parameters for cmdline tftp client
// addr - tftp server address
// port - tftp server port
// file - tftp file requested by client RRQ message
// dest - where to store requested file
type cfg struct {
	addr    string
	file    string
	dest    string
	port    int
	timeout time.Duration
}

func main() {
	config := cfg{}
	flag.StringVar(&config.addr, "addr", "", "server address")
	flag.StringVar(&config.dest, "dest", "", "file path for request file")
	flag.StringVar(&config.file, "rf", "", "filename for RRQ request")
	flag.IntVar(&config.port, "port", 69, "server port")
	flag.DurationVar(&config.timeout, "timeout", 0, "timeout for client retry, 0 no timeout")
	flag.Parse()

	if config.addr == "" {
		log.Fatal("addr should be specified")
	}
	if config.file == "" {
		log.Fatal("request file should be specified")
	}

	// if destination is not set use current folder and requested file name
	if config.dest == "" {
		config.dest = filepath.Join(".", config.file)
	}

	f, err := os.OpenFile(config.dest, os.O_CREATE|os.O_EXCL|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("can't create file to save tftp request: %v", err)
	}
	defer f.Close()

	c := &Client{}
	addr := net.JoinHostPort(config.addr, strconv.Itoa(config.port))
	err = c.Request(addr, config.file, f)
	if err != nil {
		log.Fatalf("can't request file %s: %v", config.file, err)
	}

	err = f.Sync()
	if err != nil {
		log.Fatal(err)
	}
}
