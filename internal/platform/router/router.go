package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nguyentanvinh/recommendations-be/backend/internal/platform/config"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Router struct {
	Config *config.Config
	Pool   *pgxpool.Pool
}

const prefixAPIV1 = "/api/v1"

func (r *Router) NewRouter() *mux.Router {
	router := mux.NewRouter()

	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("doc.json"),
	)).Methods(http.MethodGet)

	apiV1 := router.PathPrefix(prefixAPIV1).Subrouter()

	r.registerHealthRoute(apiV1)
	r.registerPlaceRoutes(apiV1)

	return router
}
