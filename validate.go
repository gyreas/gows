package gows

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/textproto"
	"os"
	"strings"
)

type Response struct {
	Protocol   string
	Extensions []string
}

func getResponse(response string) Response {
	fields := validateResponse(response)

	exts := []string(nil)
	if val, ok := fields["Sec-Websocket-Extension"]; ok && len(val) != 0 {
		exts = val
	}

	protocol := ""
	if val, ok := fields["Sec-Websocket-Protocol"]; ok && len(val) == 1 {
		protocol = val[0]
	}

	return Response{
		Protocol:   protocol,
		Extensions: exts,
	}
}

func validateResponse(response string) map[string][]string {
	s := bufio.NewReader(strings.NewReader(response))
	req := textproto.NewReader(s)

	statusLine, err := req.ReadLine()
	if err != nil {
		die("error: ReadLine\n")
	}
	if statusLine[:4] != "HTTP" {
		die("error: Unknown protocol: %q\n", statusLine[:4])
	}

	fields, _ := req.ReadMIMEHeader()

	upg, ok := fields["Upgrade"]
	if !ok {
		log.Fatalf("error: response has no Upgrade field\n")
	}
	if len(upg) != 1 {
		log.Fatalf("error: Upgrade field has incorrect number of values: got %d, want 1\n", len(upg))
	}
	if strings.Trim(upg[0], " ") != "websocket" {
		log.Fatalf("error: unsupported Upgrade value: %q\n", upg)
	}

	conn, ok := fields["Connection"]
	if !ok {
		log.Fatalf("error: response has no Connection field\n")
	}
	if len(conn) != 1 {
		log.Fatalf("error: Connection field has incorrect number of values: got %d, want 1\n", len(conn))
	}
	if strings.ToLower(strings.Trim(conn[0], " ")) != "upgrade" {
		log.Fatalf("error: unsupported Connection value: %q\n", conn)
	}

	accept, ok := fields["Sec-Websocket-Accept"]
	if !ok {
		log.Fatalf("error: response has no Sec-WebSocket-Accept field\n")
	}
	if len(accept) != 1 {
		log.Fatalf("error: Sec-WebSocket-Accept field has incorrect number of values: got %d, want 1\n", len(accept))
	}
	{
		sha := sha1.New()
		io.WriteString(sha, secWsKey+WS_GUID)
		data := sha.Sum(nil)

		s := strings.Builder{}
		enc := base64.NewEncoder(base64.StdEncoding, &s)
		enc.Write(data)
		enc.Close()

		if s.String() != accept[0] {
			log.Fatalf("error: Bad Response\n")
		}
	}

	return fields
}

func die(format string, arg ...any) {
	fmt.Fprintf(os.Stderr, format, arg)
	os.Exit(1)
}
