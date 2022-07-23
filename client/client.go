package client

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/binjip978/tftp/packets"
)

// Client request file over tftp protocol
type Client struct {
	Timeout time.Duration
}

// Request downloads filename from tftp server and writes to io.Writer
func (c *Client) Request(addr string, filename string, wr io.Writer) error {
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return fmt.Errorf("client can't listen on udp socket: %v", err)
	}
	defer conn.Close()

	serverAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("can't resolve udp address: %v", err)
	}

	r := packets.RRQ{Filename: filename}

	_, err = conn.WriteToUDP(r.Encode(), serverAddr)
	if err != nil {
		return fmt.Errorf("can't send rrq request: %v", err)
	}

	buffer := make([]byte, 516)
	var lastAck uint16
	retry := 3

	for {
		if c.Timeout != 0 {
			_ = conn.SetReadDeadline(time.Now().Add(c.Timeout))
		}
		n, ra, err := conn.ReadFromUDP(buffer)
		if os.IsTimeout(err) {
			retry -= 1
			if retry == 0 {
				return fmt.Errorf("can't retry more")
			}
			if lastAck > 0 {
				ack := &packets.Ack{Block: lastAck}
				_, _ = conn.WriteToUDP(ack.Encode(), ra)
			}

			continue
		}
		if err != nil {
			return fmt.Errorf("can't read from connection: %v", err)
		}
		retry = 3
		opcode := binary.BigEndian.Uint16(buffer[:2])
		switch opcode {
		case 3: // data packet, append data to file
			d, err := packets.ParseData(buffer[:n])
			if err != nil {
				return fmt.Errorf("can't parse data packet: %v", err)
			}
			_, err = wr.Write(d.Data)
			if err != nil {
				return fmt.Errorf("can't write data bytes into file: %v", err)
			}
			// send ack to the server
			ack := &packets.Ack{Block: d.Block}
			lastAck = d.Block
			_, err = conn.WriteToUDP(ack.Encode(), ra)
			if err != nil {
				return fmt.Errorf("can't write ack to the server: %v", err)
			}
			if len(d.Data) < 512 {
				return nil
			}
		case 5: // report an error to the user
			e, err := packets.ParseError(buffer[:n])
			if err != nil {
				return fmt.Errorf("can't parse error message: %v", err)
			}
			return fmt.Errorf("error: %s", e.Msg)
		default:
			return fmt.Errorf("unknown opcode from server: %d", opcode)
		}
	}
}
