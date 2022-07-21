package main

import (
	"log"
	"time"

	"github.com/binjip978/tftp/server"
)

func main() {
	handler, err := server.FileHandler("/Users/binjip978/Downloads/tiger.jpeg")
	if err != nil {
		panic(err)
	}

	srv := server.Server{
		// Handler: server.BytesHandler([]byte("hello world")),
		Handler: handler,
		Timeout: 2 * time.Second,
		Retry:   5,
	}

	log.Fatal(srv.ListenAndServe("127.0.0.1:6900"))
}
