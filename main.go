package main

import (
	"log"
	"net/http"

	pokerplan "github.com/DDimD/pokerPlaning/pokerPlaning"
)

func main() {
	log.Println("RUN")
	fileSys := http.FileServer(http.Dir("./index/"))
	http.Handle("/", fileSys)

	server := pokerplan.NewServer("/websocket")
	go server.Listen()

	log.Fatal(http.ListenAndServe(":25555", nil))
}
