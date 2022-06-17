package books

import (
	"net/http"

	"github.com/athosone/golib/pkg/server/renderer"
	"github.com/athosone/golib/pkg/server/routing"
	"github.com/go-chi/chi"
	"github.com/gorilla/mux"
)

const (
	v1Beta1BookMediaType = "application/vnd.athosone.book+*;v=v1beta1"
	v1Beta2BookMediaType = "application/vnd.athosone.book+*;    v=v1beta2"
	v2BookMediaType      = "application/vnd.athosone.book+*; v=v2"
	v1AddRating          = "application/vnd.athosone.book.rating.add+json; v=v1"
	v2BookRating         = "application/vnd.athosone.book.rating+json; v=v2"
)

func BookingRouter() *routing.GRouter {
	bookingRouter := routing.NewRouter()
	bookingRouter.Get(GetBetaBooks).Produce(v1Beta1BookMediaType, v1Beta2BookMediaType).SetDefault()
	bookingRouter.Get(GetV2Books).Produce(v2BookMediaType)

	return bookingRouter
}

func RatingRouter() *routing.GRouter {
	ratingRouter := routing.NewRouter()
	ratingRouter.Post(PostAddRatingV2).Produce(v2BookRating).Consume(v1AddRating)
	ratingRouter.Get(GetV2Ratings).Produce(v2BookRating)

	return ratingRouter
}

type V1Beta2Book struct {
	Name   string `json:"name"`
	Rating string `json:"rating"`
}

type V2Book struct {
	Name    string     `json:"fullName"`
	Ratings []V2Rating `json:"ratings"`
}

type V2Rating struct {
	Grade string `json:"grade"`
}

func GetBetaBooks(w http.ResponseWriter, r *http.Request) {
	// Long string to test content encoding
	_ = renderer.Created(w, r, V1Beta2Book{Name: "lore ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."})
}

func GetV2Books(w http.ResponseWriter, r *http.Request) {
	_ = renderer.OK(w, r, V2Book{Name: "GET V2"})
}

func GetV2Ratings(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		id = mux.Vars(r)["id"]
	}
	_ = renderer.OK(w, r, V2Rating{Grade: "get rating:" + id})
}

func PostAddRatingV2(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		id = mux.Vars(r)["id"]
	}
	_ = renderer.Created(w, r, V2Rating{Grade: "post rating: " + id})
}
