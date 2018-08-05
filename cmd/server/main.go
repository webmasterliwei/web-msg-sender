package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"html/template"
	"log"
	"net/http"
	"os"
	"encoding/json"
)

type Message struct {
	Type    string `json:"type"`
	UserId  int    `json:"user_id"`
	Content string `json:"content"`
}

var userConnections map[int]*websocket.Conn

func webSocket(ws *websocket.Conn) {
	var err error
	for {
		var received string
		//接受消息
		if err = websocket.Message.Receive(ws, &received); err != nil {
			// 用户掉线
			fmt.Println("receive failed:", err)
			fmt.Println("user connections befor delete:", userConnections)
			for userId, connection := range userConnections {
				if connection == ws {
					delete(userConnections, userId)
					break
				}
			}
			fmt.Println("user connections after delete:", userConnections)
			ws.Close()
			break
		}
		fmt.Println("received:", received)
		var message Message
		if err = json.Unmarshal([]byte(received), &message); err != nil {
			fmt.Println("json decode error:", err)
			websocket.Message.Send(ws, "json decode error: " + err.Error())
			ws.Close()
			break
		}
		fmt.Println("message:", message)
		// 消息不合法
		if !(message.Type == "login" || message.Type == "publish") {
			fmt.Println("message type error, only login or publish")
			websocket.Message.Send(ws, "message type error, only login or publish")
			ws.Close()
			break
		}
		switch message.Type {
		case "login":
			// 登录
			fmt.Println("add user connection:", &ws)
			userConnections[message.UserId] = ws
			fmt.Println("current user connections:", userConnections)
		case "publish":
			if message.UserId == 0 {
				// 发布消息给除自己外所有在线用户
				if len(userConnections) == 0 {
					websocket.Message.Send(ws, "no online user")
					ws.Close()
					return
				}
				ok := true
				for userId, userConnection := range userConnections {
					if userConnection == ws {
						continue
					}
					if err = websocket.Message.Send(userConnection, message.Content); err != nil {
						fmt.Println("publish message:", message.Content, "to user:", message.UserId, "failed:", err)
						ok = false
					}
					fmt.Println("publish message:", message.Content, "to user:", userId, "ok")
				}
				if !ok {
					websocket.Message.Send(ws, "failed")
					ws.Close()
					return
				}
				websocket.Message.Send(ws, "ok")
				ws.Close()
				return
			}
			if userConnection, ok :=  userConnections[message.UserId]; ok {
				// 发布消息给指定的用户
				if err = websocket.Message.Send(userConnection, message.Content); err != nil {
					fmt.Println("publish message:", message.Content, "to user:", message.UserId, "failed:", err)
					websocket.Message.Send(ws, "failed")
					ws.Close()
					return
				}
				fmt.Println("publish message:", message.Content, "to user:", message.UserId, "ok")
				websocket.Message.Send(ws, "ok")
				ws.Close()
				return
			}
			// 指定的用户已不在线
			fmt.Println("user offline")
			websocket.Message.Send(ws, "offline")
			ws.Close()
			return
		}
	}
}

func receiver(w http.ResponseWriter, r *http.Request) {
	wd, _ := os.Getwd()
	t, _ := template.ParseFiles(wd + "/receiver.html")
	t.Execute(w, nil)
}

func publisher(w http.ResponseWriter, r *http.Request) {
	wd, _ := os.Getwd()
	t, _ := template.ParseFiles(wd + "/publisher.html")
	t.Execute(w, nil)
}

func main() {
	userConnections = make(map[int]*websocket.Conn)
	http.Handle("/webSocket", websocket.Handler(webSocket))
	http.HandleFunc("/receiver", receiver)
	http.HandleFunc("/publisher", publisher)
	if err := http.ListenAndServe(":1234", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
