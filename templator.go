package main

import (
	"net/http"

	pokerplan "github.com/DDimD/pokerPlaning/pokerPlaning"
)

type ViewData struct {
	Users []User
	cards []float64
	Topic string
}

func templateHandler(w http.ResponseWriter, r *http.Request, srv *pokerplan.Server) {
	users := srv.
	vievData := ViewData{

	}
}
