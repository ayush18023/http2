package main

import (
	"io"
)

type StreamIDType = uint32
type Connection struct {
	rw      io.ReadWriter
	streams map[StreamIDType]*Stream
}

func (c *Connection) readLoop(r io.Reader) {
	// header := new(Headers)
	for {

	}
}
