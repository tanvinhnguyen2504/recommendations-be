package router

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (r *Router) registerHealthRoute(router *mux.Router) {
	router.HandleFunc("/", r.healthCheck).Methods(http.MethodGet)
	router.HandleFunc("/health", r.healthCheck).Methods(http.MethodGet)
}

func (r *Router) healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	payload := map[string]string{"env": r.Config.Env, "image": r.Config.Image}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("router: health response: %v", err)
	}
}
