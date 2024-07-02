package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"pokerPlaning/pokerplan"
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

type card struct {
	Value      float64
	IsCoffee   bool
	IsQuestion bool
}

type ViewData struct {
	Users []pokerplan.User
	Cards []card
	Topic string
}

// func templateHandler(w http.ResponseWriter, r *http.Request) {
// 	viewData := ViewData{
// 		srv.GetClients(),
// 		prepareCards(),
// 		srv.GetTopicName(),
// 	}

// 	tmpl, _ := template.ParseFiles("templates/index.html")
// 	tmpl.Execute(w, viewData)
// }

func prepareCards() []card {
	cards := make([]card, 0, 12)
	var cache [2]float64
	cards = append(cards, card{0., false, false})
	cards = append(cards, card{0.5, false, false})
	cache[0] = 0
	cache[1] = 1

	for i := 0; i < 8; i++ {
		cache[i%2] = cache[0] + cache[1]
		cards = append(cards, card{cache[i%2], false, false})
	}

	cards = append(cards, card{0., true, false})
	cards = append(cards, card{0., false, true})
	return cards
}
