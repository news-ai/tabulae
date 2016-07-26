package router

import (
	"net/http"

	"github.com/news-ai/tabulae/routes"
)

type Handler func(http.ResponseWriter, *http.Request)
type Resource map[string]Handler
type Action struct {
	HandlerName string
	Routes      Resource
}

func GetRoutes() []Action {
	router := []Action{
		Action{
			"/users", map[string]Handler{
				"/":     routes.UsersHandler,
				"/{id}": routes.UserHandler,
			},
		},
		Action{
			"/agencies", map[string]Handler{
				"/":     routes.AgenciesHandler,
				"/{id}": routes.AgencyHandler,
			},
		},
		Action{
			"/publications", map[string]Handler{
				"/":     routes.PublicationsHandler,
				"/{id}": routes.PublicationHandler,
			},
		},
		Action{
			"/contacts", map[string]Handler{
				"/":     routes.ContactsHandler,
				"/{id}": routes.ContactHandler,
			},
		},
		Action{
			"/lists", map[string]Handler{
				"/":              routes.MediaListsHandler,
				"/{id}":          routes.MediaListHandler,
				"/{id}/{action}": routes.UploadHandler,
			},
		},
		Action{
			"/files", map[string]Handler{
				"/":     routes.FilesHandler,
				"/{id}": routes.FileHandler,
			},
		},
	}
	return router
}
