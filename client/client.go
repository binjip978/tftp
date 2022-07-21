package main

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/binjip978/tftp/packets"
)

// Client request file over tftp protocol
type Client struct {
	Timeout time.Duration
}

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
	var lastAck uint16

	for {
		if c.Timeout != 0 {
			_ = conn.SetReadDeadline(time.Now().Add(c.Timeout))
		}
		n, ra, err := conn.ReadFromUDP(buffer)
		if os.IsTimeout(err) {
			ack := &packets.Ack{Block: lastAck}
			_, err = conn.WriteToUDP(ack.Encode(), ra)
			if err != nil {
				log.Fatalf("can't write ack to the server: %v", err)
			}
			continue
		}
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
			lastAck = d.Block
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
