package packets

import (
	"bytes"
	"testing"
)

func TestRRQ(t *testing.T) {
	r0 := RRQ{Filename: "hello.txt"}
	b := r0.Encode()

	r1, err := ParseRRQ(b)
	if err != nil {
		t.Error(err)
	}

	if r1.Filename != r0.Filename {
		t.Error("filename is not the same after encode/decode")
	}
}

func TestData(t *testing.T) {
	d0 := Data{
		Block: 1,
		Data:  []byte("hello world"),
	}

	b := d0.Encode()
	d1, err := ParseData(b)
	if err != nil {
		panic(err)
	}

	if d1.Block != d0.Block {
		t.Error("block number is not the same after encode/decode")
	}
	if !bytes.Equal(d1.Data, d0.Data) {
		t.Error("data is not the same after encode/decode")
	}
}

func TestAck(t *testing.T) {
	a0 := Ack{Block: 13}
	b := a0.Encode()
	a1, err := ParseAck(b)
	if err != nil {
		t.Error(err)
	}

	if a0.Block != a1.Block {
		t.Error("ack block is not the same after encode/decode")
	}
}

func TestError(t *testing.T) {
	e0 := Error{
		Code: 1,
		Msg:  "File not found.",
	}

	b := e0.Encode()
	e1, err := ParseError(b)
	if err != nil {
		t.Error(err)
	}

	if e0.Code != e1.Code {
		t.Error("error code is not the same after encode/decode")
	}
	if e0.Msg != e1.Msg {
		t.Error("message is not the same after encode/decode")
	}
}
