package main

import (
	"encoding/binary"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/binjip978/tftp/packets"
)

// cfg defines parameters for cmdline tftp client
// addr - tftp server address
// port - tftp server port
// file - tftp file requested by client RRQ message
// dest - where to store requested file
type cfg struct {
	addr string
	file string
	dest string
	port int
}

// Client request file over tftp protocol
// TODO: add retry params
type Client struct{}

func (c *Client) Request(addr string, filename string, wr io.Writer) error {
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		log.Fatalf("client can't listen on udp socket: %v", err)
	}
	defer conn.Close()

	serverAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatalf("can't resolve udp address: %v", err)
	}

	r := packets.RRQ{Filename: filename}

	_, err = conn.WriteToUDP(r.Encode(), serverAddr)
	if err != nil {
		log.Fatalf("can't send rrq request: %v", err)
	}

	buffer := make([]byte, 516)

	for {
		n, ra, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Fatalf("can't read from connection: %v", err)
		}
		opcode := binary.BigEndian.Uint16(buffer[:2])
		switch opcode {
		case 3: // data packet, append data to file
			d, err := packets.ParseData(buffer[:n])
			if err != nil {
				log.Fatalf("can't parse data packet: %v", err)
			}
			_, err = wr.Write(d.Data)
			if err != nil {
				log.Fatalf("can't write data bytes into file: %v", err)
			}
			// send ack to the server
			ack := &packets.Ack{Block: d.Block}
			_, err = conn.WriteToUDP(ack.Encode(), ra)
			if err != nil {
				log.Fatalf("can't write ack to the server: %v", err)
			}
			if len(d.Data) < 512 {
				return nil
			}
		case 5: // report an error to the user
			e, err := packets.ParseError(buffer[:n])
			if err != nil {
				log.Fatalf("can't parse error message: %v", err)
			}
			log.Fatal(e.Msg)
		default:
			log.Fatalf("unknown opcode from server: %d", opcode)
		}
	}
}

func main() {
	config := cfg{}
	flag.StringVar(&config.addr, "addr", "", "server address")
	flag.StringVar(&config.dest, "dest", "", "file path for request file")
	flag.StringVar(&config.file, "rf", "", "filename for RRQ request")
	flag.IntVar(&config.port, "port", 69, "server port")
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
