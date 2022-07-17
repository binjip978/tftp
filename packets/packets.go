// Package packets implements encoding/decoding of TFTP packets
// See RFC 1350 https://datatracker.ietf.org/doc/html/rfc1350

package packets

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// octet is only supported formant, packets with netascii
// and mail should be dropped
const octet = "octet"

// RRQ is tftp request for file
type RRQ struct {
	Filename string
}

// Encode RRQ to bytes
func (r *RRQ) Encode() []byte {
	buf := bytes.Buffer{}
	buf.WriteByte(0)
	buf.WriteByte(1)
	buf.WriteString(r.Filename)
	buf.WriteByte(0)
	buf.WriteString(octet)
	buf.WriteByte(0)
	return buf.Bytes()
}

// ParseRRQ parse bytes to RRQ
func ParseRRQ(b []byte) (*RRQ, error) {
	opcode := binary.BigEndian.Uint16(b[:2])
	if opcode != 1 {
		return nil, fmt.Errorf("opcode is not 1")
	}

	filename, b, err := readTFTPString(b[2:])
	if err != nil {
		return nil, err
	}

	r := &RRQ{Filename: filename}

	mode, b, err := readTFTPString(b)
	if err != nil {
		return nil, err
	}

	if len(b) != 0 {
		return nil, fmt.Errorf("more bytes than expected")
	}

	if mode != octet {
		return nil, fmt.Errorf("octet mode is only supported")
	}

	return r, nil
}

func readTFTPString(b []byte) (string, []byte, error) {
	cnt := 0

	for cnt < len(b) {
		if b[cnt] == 0 {
			return string(b[0:cnt]), b[cnt+1:], nil
		}
		cnt++
	}

	return "", nil, fmt.Errorf("no 0 byte found")
}

// Data is tftp DATA packet
type Data struct {
	Data  []byte
	Block uint16
}

// Encode DATA to bytes
func (d *Data) Encode() []byte {
	b := make([]byte, 2+2+len(d.Data))
	b[0] = 0
	b[1] = 3
	binary.BigEndian.PutUint16(b[2:], d.Block)
	copy(b[4:], d.Data)
	return b
}

// ParseData parse bytes to DATA
func ParseData(b []byte) (*Data, error) {
	opcode := binary.BigEndian.Uint16(b[:2])
	if opcode != 3 {
		return nil, fmt.Errorf("opcode is not 3")
	}

	d := &Data{}
	d.Block = binary.BigEndian.Uint16(b[2:4])
	d.Data = b[4:]

	return d, nil
}

// Ack is tftp ACK packet
type Ack struct {
	Block uint16
}

// Encode ACK to bytes
func (a *Ack) Encode() []byte {
	b := make([]byte, 4)
	b[0] = 0
	b[1] = 4
	binary.BigEndian.PutUint16(b[2:], a.Block)
	return b
}

// ParseAck parse bytes to ACK
func ParseAck(b []byte) (*Ack, error) {
	opcode := binary.BigEndian.Uint16(b[:2])
	if opcode != 4 {
		return nil, fmt.Errorf("opcode is not 4")
	}

	a := &Ack{}
	a.Block = binary.BigEndian.Uint16(b[2:])

	return a, nil
}

// Error is tftp ERROR packet
type Error struct {
	Msg  string
	Code uint16
}

// Encode ERROR to bytes
func (e *Error) Encode() []byte {
	b := make([]byte, 2+2+len(e.Msg)+1)
	b[0] = 0
	b[1] = 5
	binary.BigEndian.PutUint16(b[2:], e.Code)
	copy(b[4:], []byte(e.Msg))
	b[len(b)-1] = 0
	return b
}

// ParseError parse bytes to ERROR
func ParseError(b []byte) (*Error, error) {
	opcode := binary.BigEndian.Uint16(b[:2])
	if opcode != 5 {
		return nil, fmt.Errorf("opcode is not 5")
	}

	e := &Error{}
	e.Code = binary.BigEndian.Uint16(b[2:4])

	msg, b, err := readTFTPString(b[4:])
	if err != nil {
		return nil, err
	}

	if len(b) != 0 {
		return nil, fmt.Errorf("more bytes than expected")
	}

	e.Msg = msg
	return e, nil
}
