package main

import (
	"fmt"
	"flag"
	"golang.org/x/net/websocket"
	"log"
	"encoding/json"
	"strconv"
)

func main() {
	userIdStr := flag.String("user_id", "0", "user id")
	content := flag.String("content", "hello world!", "content")
	flag.Parse()
	ws, err := websocket.Dial("ws://127.0.0.1:1234/webSocket", "", "ws://127.0.0.1:1234")
	if err != nil {
		log.Fatal(err)
	}
	userId, err := strconv.Atoi(*userIdStr)
	if err != nil {
		log.Fatal(err)
	}
	message, err := json.Marshal(map[string]interface{}{
		"type": "publish",
		"user_id": userId,
		"content": *content,
	})
	if err != nil {
		log.Fatal(err)
	}
	if err = websocket.Message.Send(ws, string(message)); err != nil {
		log.Fatal(err)
	}
	var received string
	if err = websocket.Message.Receive(ws, &received); err != nil {
		log.Fatal(err)
	}
	fmt.Println(received)
}