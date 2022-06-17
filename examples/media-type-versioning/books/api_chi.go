package books

import (
	"github.com/go-chi/chi/v5"
)

func SetupWithChi(router chi.Router) {
	router.Route("/books", func(r chi.Router) {
		r.Mount("/", BookingRouter())
		r.Route("/{id}/ratings", func(r chi.Router) {
			r.Mount("/", RatingRouter())
		})
	})
}
