package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"
)

const bufferSize = 4096

type Connection interface {
	Close() error
}

type WebSocket struct {
	conn   Connection
	bufrw  *bufio.ReadWriter
	header http.Header
	status uint16
}

// hijacking HTTP connection and return WebSocket
func New(w http.ResponseWriter, r *http.Request) (*WebSocket, error) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, errors.New("webserver doesn't support http hijacking")
	}
	conn, bufrw, err := hj.Hijack()
	if err != nil {
		return nil, err
	}

	return &WebSocket{conn, bufrw, r.Header, 1000}, nil
}

func getAcceptHash(key string) string {
	h := sha1.New()
	h.Write([]byte(key))
	h.Write([]byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (ws *WebSocket) Handshake() error {
	hash := getAcceptHash(ws.header.Get("Sec-WebSocket-Key"))
	lines := []string{
		"HTTP/1.1 101 Web Socket Protocol Handshake",
		"Server: go/echoserver",
		"Upgrade: websocket",
		"Connection: Upgrade",
		"Sec-WebSocket-Accept: " + hash,
		"", // required for extra CRLF
		"", // required for extra CRLF
	}

	return ws.write([]byte(strings.Join(lines, "\r\n")))

}

func (ws *WebSocket) read(size int) ([]byte, error) {
	data := make([]byte, 0)
	for {
		if len(data) == size {
			break
		}
		sz := bufferSize
		remaining := size - len(data)
		if sz > remaining {
			sz = remaining
		}

		temp := make([]byte, sz)

		n, err := ws.bufrw.Read(temp)
		if err != nil && err != io.EOF {
			return data, err
		}

		data = append(data, temp[:n]...)
	}

	return data, nil
}

func (ws *WebSocket) write(data []byte) error {
	if _, err := ws.bufrw.Write(data); err != nil {
		return err
	}
	return ws.bufrw.Flush()
}

func (ws *WebSocket) Send(fr Frame) error {
	data := make([]byte, 2)
	data[0] = fr.Opcode
	// Basic payload with payload length <= 125
	data[1] = byte(len(fr.PayloadData))
	data = append(data, fr.PayloadData...)

	return ws.write(data)
}

func (ws *WebSocket) Recv() (Frame, error) {
	frame := Frame{}
	head, err := ws.read(2)
	if err != nil {
		return frame, err
	}

	frame.Opcode = head[0]
	length := uint64(head[1] & 0x7F)

	mask, err := ws.read(4)
	if err != nil {
		return frame, err
	}

	frame.PayloadLength = length
	payload, err := ws.read(int(length))

	if err != nil {
		return frame, err
	}

	for i := uint64(0); i < length; i++ {
		payload[i] ^= mask[i%4]
	}
	frame.PayloadData = payload
	return frame, nil
}
