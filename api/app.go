package tabulae

import (
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/news-ai/tabulae/middleware"
	"github.com/news-ai/tabulae/router"
)

func init() {
	// Setting up Negroni Router
	app := negroni.New()
	app.Use(negroni.NewRecovery())
	app.Use(negroni.NewLogger())
	app.Use(negroni.HandlerFunc(middleware.UpdateOrCreateUser))

	// CORs
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://newsai.org", "http://localhost:3000"},
		AllowCredentials: true,
	})
	app.Use(c)

	// API router
	api := mux.NewRouter().PathPrefix("/api").Subrouter().StrictSlash(true)

	// Register routes
	apiRoutes := router.GetRoutes()
	for i := 0; i < len(apiRoutes); i++ {
		api.HandleFunc(apiRoutes[i].HandlerName, apiRoutes[i].Routes["/"])
		if len(apiRoutes[i].Routes) > 1 {
			routeName := "/{id}"
			api.HandleFunc(apiRoutes[i].HandlerName+routeName, apiRoutes[i].Routes[routeName])
		}
	}

	// Main router
	main := mux.NewRouter().StrictSlash(true)
	main.PathPrefix("/api").Handler(negroni.New(negroni.Wrap(api)))

	// HTTP router
	app.UseHandler(main)
	http.Handle("/", context.ClearHandler(app))
}
