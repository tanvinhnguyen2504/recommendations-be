package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nguyentanvinh/recommendations-be/backend/internal/places"
	"github.com/nguyentanvinh/recommendations-be/backend/internal/platform/handler"
)

func (r *Router) registerPlaceRoutes(router *mux.Router) {
	repo := places.NewRepository(r.Pool)
	handler := places.NewHandler(handler.New(r.Config), places.NewService(repo))

	router.HandleFunc("/places", handler.ListPlaces).Methods(http.MethodGet)
	router.HandleFunc("/places", handler.CreatePlace).Methods(http.MethodPost)
	router.HandleFunc("/places/{id:[0-9]+}", handler.GetPlaceByID).Methods(http.MethodGet)
	router.HandleFunc("/places/{id:[0-9]+}", handler.UpdatePlace).Methods(http.MethodPut)
	router.HandleFunc("/places/{id:[0-9]+}", handler.DeletePlace).Methods(http.MethodDelete)
}
