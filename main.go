package main

import (
	"log"
	"net/http"

	"github.com/DDimD/chatServer/chat"
)

func main() {
	log.Println("RUN")
	fileSys := http.FileServer(http.Dir("./index/"))
	http.Handle("/", fileSys)

	server := chat.NewServer("/websocket")
	go server.Listen()

	log.Fatal(http.ListenAndServe(":25555", nil))
}
