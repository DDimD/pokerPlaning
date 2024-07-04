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

	cards = append(cards, card{0., false, false})
	cards = append(cards, card{0.5, false, false})
	cards = append(cards, card{1, false, false})
	cards = append(cards, card{2, false, false})
	cards = append(cards, card{4, false, false})
	cards = append(cards, card{6, false, false})
	cards = append(cards, card{8, false, false})
	cards = append(cards, card{10, false, false})
	cards = append(cards, card{12, false, false})
	cards = append(cards, card{16, false, false})
	cards = append(cards, card{20, false, false})
	cards = append(cards, card{24, false, false})
	cards = append(cards, card{28, false, false})
	cards = append(cards, card{32, false, false})
	cards = append(cards, card{40, false, false})
	cards = append(cards, card{0., true, false})
	cards = append(cards, card{0., false, true})
	return cards
}
