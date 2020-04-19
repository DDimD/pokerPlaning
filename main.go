package main

import (
	"html/template"
	"log"
	"net/http"

	pokerplan "github.com/DDimD/pokerPlaning/pokerPlaning"
	"github.com/gorilla/mux"
)

func main() {
	log.Println("RUN")

	server := pokerplan.NewServer("/websocket")
	go server.Listen()

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		viewData := ViewData{
			server.GetClients(),
			prepareCards(),
			server.GetTopicName(),
		}

		tmpl, _ := template.ParseFiles("index/index.html")
		tmpl.Execute(w, viewData)
	})

	r.PathPrefix("/").Handler(http.StripPrefix("/",
		http.FileServer(http.Dir("./index/"))))

	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":25555", nil))
}
