package dev

import (
	"fmt"
	"io"
	"strings"
)

type PortWriter struct {
	buffer io.Writer
	end    chan bool
	len    int
}

// NewPortWriter creates a writer that will write the content in buffer,
// and Wait will return after strings "Forwarding from 127.0.0.1:" has been written "len" times
func NewPortWriter(buffer io.Writer, len int) *PortWriter {
	return &PortWriter{
		buffer: buffer,
		len:    len,
		end:    make(chan bool),
	}
}

func (o *PortWriter) Write(buf []byte) (n int, err error) {
	s := string(buf)
	if strings.HasPrefix(s, "Forwarding from 127.0.0.1") {
		fmt.Fprintf(o.buffer, " - %s", s)
		o.len--
		if o.len == 0 {
			o.end <- true
		}
	}
	return len(buf), nil
}

func (o *PortWriter) Wait() {
	<-o.end
}
