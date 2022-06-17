package books

import (
	"github.com/gorilla/mux"
)

func SetupMux(router *mux.Router) {
	router.Handle("/books", BookingRouter())
	router.Handle("/books/{id}/ratings", RatingRouter())
}
