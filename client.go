package gows

import (
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

func (client *WebSocketClient) Send(data []byte) error {
	if _, err := client.conn.Write(data); err != nil {
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
	opFIN   op = 0x0
	opTEXT     = 0x1
	opBIN      = 0x2
	opCLOSE    = 0x8
	opPING     = 0x9
	opPONG     = 0xa
)

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
