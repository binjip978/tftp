package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/binjip978/tftp/packets"
)

var errFileNotFound = errors.New("file not found")

type Handler func(path string) (io.ReadCloser, error)

type Server struct {
	Handler Handler
	Timeout time.Duration
	Retry   int
}

// FileHandler returns Handler that serve from file system path
func FileHandler(path string) (Handler, error) {
	return func(rf string) (io.ReadCloser, error) {
		if filepath.Base(path) != rf {
			return nil, errFileNotFound
		}

		_, err := os.Stat(path)
		if err != nil {
			return nil, errFileNotFound
		}

		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}

		return f, err
	}, nil
}

// BytesHandler returns Handler that serve specified bytes
func BytesHandler(b []byte) Handler {
	r := bytes.NewReader(b)
	c := io.NopCloser(r)

	return func(_ string) (io.ReadCloser, error) {
		return c, nil
	}
}

func (s *Server) ListenAndServe(addr string) error {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}

	// maximum length of filename is 255 bytes, so buffer
	// size if 2(opcode) + 255(filename) + 1(0) + 5(octet) + 1(0)
	b := make([]byte, 264)

	for {
		n, clientAddr, err := conn.ReadFrom(b)
		if err != nil {
			continue
		}

		rrq, err := packets.ParseRRQ(b[:n])
		if err != nil {
			continue
		}

		go s.transfer(rrq, clientAddr)
	}
}

func (s *Server) transfer(rrq *packets.RRQ, addr net.Addr) {
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return
	}
	defer conn.Close()

	r, err := s.Handler(rrq.Filename)
	if errors.Is(err, errFileNotFound) {
		e := packets.Error{Msg: err.Error(), Code: 1}
		_, _ = conn.WriteTo(e.Encode(), addr)
		return
	}
	if err != nil {
		e := packets.Error{Msg: err.Error(), Code: 0}
		_, _ = conn.WriteTo(e.Encode(), addr)
		return
	}
	defer r.Close()

	dataBuf := make([]byte, 512)
	ackBuf := make([]byte, 4) // opcode(2) + block(2)
	var block uint16 = 1

	for {
		nr, err := r.Read(dataBuf)
		if err != nil {
			e := packets.Error{Msg: err.Error(), Code: 0}
			_, _ = conn.WriteTo(e.Encode(), addr)
		}

		data := packets.Data{Data: dataBuf[0:nr], Block: block}

		acked := false
		for i := 0; i < s.Retry; i++ {
			_, err = conn.WriteTo(data.Encode(), addr)
			if err != nil {
				e := packets.Error{Msg: err.Error(), Code: 0}
				_, _ = conn.WriteTo(e.Encode(), addr)
			}

			_ = conn.SetReadDeadline(time.Now().Add(s.Timeout))
			na, err := conn.Read(ackBuf)
			if os.IsTimeout(err) {
				continue
			}
			if err != nil {
				e := packets.Error{Msg: err.Error(), Code: 0}
				_, _ = conn.WriteTo(e.Encode(), addr)
			}

			ack, err := packets.ParseAck(ackBuf[0:na])
			if err != nil {
				e := packets.Error{Msg: err.Error(), Code: 0}
				_, _ = conn.WriteTo(e.Encode(), addr)
			}

			if ack.Block == block {
				block++
				acked = true
				break
			}

			e := packets.Error{Msg: fmt.Sprintf("unexpected block id: %d", ack.Block), Code: 0}
			_, _ = conn.WriteTo(e.Encode(), addr)
		}

		if !acked {
			e := packets.Error{Msg: "can't retry more", Code: 0}
			_, _ = conn.WriteTo(e.Encode(), addr)

		}

		if nr < 512 {
			return
		}
	}
}
