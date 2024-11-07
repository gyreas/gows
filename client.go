// MIT License Copyright (c) 2024 Saheed Adeleye [aadesaed <@> gmail <.> com]

package gows

import (
	"crypto/rand"
	"fmt"
	"net"
	"net/url"
	"strings"
)

type Conn net.Conn

const (
	WS_VERSION = uint8(13)
	WS_GUID    = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	secWsKey   = "dGhlIHNhbXBsZSBub25jZQ=="
)

type WebSocketClient struct {
	conn      Conn
	host      string
	port      string
	protocols []string

	key string
}

type PayloadKind uint8

const (
	PlKindText PayloadKind = iota
	PlKindBinary
)

func NewWebSocketClient(url string /*protocols []string*/) (WebSocketClient, error) {
	client := WebSocketClient{
		protocols: []string{"chat"},
		key:       "dGhlIHNhbXBsZSBub25jZQ==",
	}
	err := connect(&client, url)

	return client, err
}

func connect(client *WebSocketClient, addr string) error {
	url, err := url.Parse(addr)
	if err != nil {
		return fmt.Errorf("invalid address %q: %s", addr, err.Error())
	}

	urlString := url.Hostname() + ":" + url.Port()
	srvAddr, err := net.ResolveTCPAddr("tcp", urlString)
	if err != nil {
		return fmt.Errorf("couldnot resolve address %q: %s", urlString, err.Error())
	}
	client.host = url.Hostname()
	client.port = url.Port()

	conn, err := net.DialTCP("tcp", nil, srvAddr)
	if err != nil {
		return fmt.Errorf("couldnot establish connection to %q: %s", urlString, err.Error())
	}

	handshake := []byte(client.handshake())
	_, err = conn.Write(handshake)
	if err != nil {
		return fmt.Errorf("client handshake failed: %s", err.Error())
	}

	buf := [1024]byte{}
	n, err := conn.Read(buf[:])
	if err != nil {
		return fmt.Errorf("no server handshake: %s", err.Error())
	}
	response := getResponse(string(buf[:n]))
	fmt.Println("Handshake successful", response)

	client.conn = conn
	return nil
}

func (client *WebSocketClient) SendText(text string) error {
	return client.send([]byte(text), PlKindText)
}

func (client *WebSocketClient) SendBlob(data []byte) error {
	return client.send(data, PlKindBinary)
}

func (client *WebSocketClient) send(data []byte, kind PayloadKind) error {
	payload := client.makePayload(data, kind)
	if _, err := client.conn.Write(payload); err != nil {
		return err
	}
	return nil
}

func (client *WebSocketClient) Read(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}

	n, err := client.conn.Read(buf)
	if err != nil {
		return -1, err
	}
	return n, nil
}

func (client *WebSocketClient) Close() error {
	err := client.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

type op uint8

const (
	opCONT  op = 0x0 /* continuation frame */
	opTEXT     = 0x1 /* text frame */
	opBIN      = 0x2 /* binary frame */
	opCLOSE    = 0x8 /* connection close */
	opPING     = 0x9 /* ping */
	opPONG     = 0xa /* pong */
)

func (client *WebSocketClient) makePayload(data []byte, payloadKind PayloadKind) []byte {
	payload := []byte{}

	dataLen := len(data)
	if dataLen < 126 {
		kind := opBIN
		if payloadKind == PlKindText {
			kind = opTEXT
		}

		frameMeta := uint8((1 << 7) | kind)
		payload = append(payload, frameMeta, uint8((1<<7)|dataLen))
	} else if dataLen == 126 {
	} else if dataLen == 127 {
	}
	mask := generateMask()
	payload = append(payload, mask[:]...)

	// mask the data
	for i, octet := range data {
		j := i % 4
		payload = append(payload, octet^mask[j])
	}

	return payload
}

func generateMask() [4]byte {
	mask := [4]byte{}
	if _, err := rand.Read(mask[:]); err != nil {
		die("internal error: couldnot generate mask: %s\n", err.Error())
	}

	return mask
}

func (client *WebSocketClient) handshake() string {
	return fmt.Sprintf(""+
		"GET / HTTP/1.1\r\n"+
		"Host: %s\r\n"+
		"Connection: Upgrade\r\n"+
		"Upgrade: websocket\r\n"+
		"Sec-Websocket-Key: %s\r\n"+
		"Sec-Websocket-Protocol: %s\r\n"+
		"Sec-Websocket-Version: %d\r\n"+
		"\r\n",
		client.host, client.key, strings.Join(client.protocols, ", "), WS_VERSION,
	)
}
