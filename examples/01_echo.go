package main

import (
	"fmt"
	ws "github.com/gyreas/gows"
	"os"
)

func main() {
	client, err := ws.NewWebSocketClient("ws://localhost:4440")
	if err != nil {
		fmt.Println(err.Error())
	}

	if err := client.SendText("'Twelve'"); err != nil {
		die(err.Error())
	}
}

func die(format string, arg ...any) {
	fmt.Fprintf(os.Stderr, format, arg)
	os.Exit(1)
}
