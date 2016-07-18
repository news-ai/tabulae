package tabulae

import (
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/news-ai/tabulae/middleware"
	"github.com/news-ai/tabulae/routes"
)

func init() {
	// Setting up Negroni Router
	app := negroni.New()
	app.Use(negroni.NewRecovery())
	app.Use(negroni.NewLogger())
	app.Use(negroni.HandlerFunc(middleware.UpdateOrCreateUser))

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
	api.HandleFunc("/agencies", routes.AgenciesHandler)
	api.HandleFunc("/agencies/{id}", routes.AgencyHandler)
	api.HandleFunc("/publications", routes.PublicationsHandler)
	api.HandleFunc("/publications/{id}", routes.PublicationHandler)
	api.HandleFunc("/contacts", routes.ContactsHandler)
	api.HandleFunc("/contacts/{id}", routes.ContactHandler)
	// api.HandleFunc("/lists", routes.ListsHandler)

	// Main router
	main := mux.NewRouter().StrictSlash(true)
	main.PathPrefix("/api").Handler(negroni.New(negroni.Wrap(api)))

	// HTTP router
	app.UseHandler(main)
	http.Handle("/", context.ClearHandler(app))
}
