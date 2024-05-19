package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

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

func (ws *WebSocket) write(data []byte) error {
	if _, err := ws.bufrw.Write(data); err != nil {
		return err
	}
	return ws.bufrw.Flush()
}
