package stream

import (
	"bytes"
	"encoding/binary"
	"io"
)

// Read consumes count bytes from the given connection and returns them
//func Read(conn net.Conn, count int) []byte {
func Read(r io.Reader, count int) []byte {
	b := make([]byte, count, count)
	// Read enough bytes for the message header
	if count == 0 {
		return nil
	}
	for {
		n, err := r.Read(b)
		if err != nil {
			//f.sendEvent(tcpConnectionFails) // This should not be here
		}
		if n < count {
			continue
		}
		break
	}
	return b[:count]
}

// ReadBytes reads n bytes from the byte buffer and returns it
func ReadBytes(n int, buf *bytes.Buffer) []byte {
	bs := make([]byte, n, n)
	for i := range bs {
		bs[i], _ = buf.ReadByte()
	}
	return bs
}

// ReadByte reads a single byte off the given byte buffer and returns it
func ReadByte(buf *bytes.Buffer) byte {
	return ReadBytes(1, buf)[0]
}

// ReadUint16 reads 2 bytes off the buffer and returns it as a uint16
func ReadUint16(buf *bytes.Buffer) uint16 {
	return binary.BigEndian.Uint16(ReadBytes(2, buf))
}

// ReadUint32 reads 4 bytes off the buffer and returns it as a uint32
func ReadUint32(buf *bytes.Buffer) uint32 {
	return binary.BigEndian.Uint32(ReadBytes(4, buf))
}
