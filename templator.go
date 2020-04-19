package main

import (
	pokerplan "github.com/DDimD/pokerPlaning/pokerPlaning"
)

type card struct {
	Value      float64
	IsCoffee   bool
	IsQuestion bool
}

type ViewData struct {
	Users []pokerplan.Client
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
