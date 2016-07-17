package tabulae

import (
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/news-ai/tabulae/routes"
)

func init() {
	// Setting up Negroni Router
	app := negroni.New()
	app.Use(negroni.NewRecovery())
	app.Use(negroni.NewLogger())

	// CORs
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"https://newsai.org"},
	})
	app.Use(c)

	// API router
	api := mux.NewRouter().PathPrefix("/api").Subrouter().StrictSlash(true)
	api.HandleFunc("/", routes.BaseHandler)
	api.HandleFunc("/users", routes.UsersHandler)
	api.HandleFunc("/users/{id}", routes.UserHandler)
	// api.HandleFunc("/lists", routes.ListsHandler)
	// api.HandleFunc("/agencies", routes.AgenciesHandler)

	// Main router
	main := mux.NewRouter().StrictSlash(true)
	main.PathPrefix("/api").Handler(negroni.New(negroni.Wrap(api)))

	// HTTP router
	app.UseHandler(main)
	http.Handle("/", context.ClearHandler(app))
}
