package tabulae

import (
	"net/http"
	"os"

	"github.com/bradleyg/go-sentroni"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
	"github.com/gorilla/csrf"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"github.com/unrolled/secure"

	"github.com/news-ai/tabulae/auth"
	"github.com/news-ai/tabulae/incoming"
	"github.com/news-ai/tabulae/middleware"
	"github.com/news-ai/tabulae/routes"
	"github.com/news-ai/tabulae/utils"
)

func init() {
	// Setting up Negroni Router
	app := negroni.New()
	app.Use(negroni.NewRecovery())
	app.Use(negroni.NewLogger())

	// CORs
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://newsai.org", "http://localhost:3000", "https://site.newsai.org", "http://site.newsai.org"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		Debug:            true,
		AllowedHeaders:   []string{"*"},
	})
	app.Use(c)

	// Initialize CSRF
	CSRF := csrf.Protect([]byte(os.Getenv("CSRFKEY")), csrf.Secure(false))

	// Initialize the environment for a particular URL
	utils.InitURL()
	auth.SetRedirectURL()

	// Initialize router
	router := httprouter.New()

	// Not found Handler
	router.GET("/", routes.NotFoundHandler)
	router.GET("/api", routes.NotFoundHandler)

	/*
	* Authentication Handler
	 */

	router.Handler("GET", "/api/auth", CSRF(auth.PasswordLoginPageHandler()))
	router.Handler("GET", "/api/auth/forget", CSRF(auth.ForgetPasswordPageHandler()))
	router.Handler("GET", "/api/auth/confirmation", CSRF(auth.EmailConfirmationHandler()))
	router.Handler("GET", "/api/auth/registration", CSRF(auth.PasswordRegisterPageHandler()))
	router.Handler("POST", "/api/auth/userlogin", CSRF(auth.PasswordLoginHandler()))
	router.Handler("POST", "/api/auth/userregister", CSRF(auth.ForgetPasswordHandler()))
	router.Handler("POST", "/api/auth/userforget", CSRF(auth.ForgetPasswordHandler()))

	router.GET("/api/auth/google", auth.GoogleLoginHandler)
	router.GET("/api/auth/callback", auth.GoogleCallbackHandler)

	router.GET("/api/auth/logout", auth.LogoutHandler)

	/*
	* Incoming Handler
	 */

	router.POST("/api/incoming/sendgrid", incoming.SendGridHandler)

	/*
	* API Handler
	 */

	router.GET("/api/users", routes.UsersHandler)
	router.GET("/api/users/:id", routes.UserHandler)
	router.PATCH("/api/users/:id", routes.UserHandler)

	router.GET("/api/agencies", routes.AgenciesHandler)
	router.GET("/api/agencies/:id", routes.AgencyHandler)

	router.GET("/api/publications", routes.PublicationsHandler)
	router.POST("/api/publications", routes.PublicationsHandler)
	router.GET("/api/publications/:id", routes.PublicationHandler)

	router.GET("/api/contacts", routes.ContactsHandler)
	router.POST("/api/contacts", routes.ContactsHandler)
	router.PATCH("/api/contacts", routes.ContactsHandler)
	router.GET("/api/contacts/:id", routes.ContactHandler)
	router.PATCH("/api/contacts/:id", routes.ContactHandler)
	router.GET("/api/contacts/:id/:action", routes.ContactActionHandler)

	router.GET("/api/files", routes.FilesHandler)
	router.GET("/api/files/:id", routes.FileHandler)
	router.GET("/api/files/:id/:action", routes.FileActionHandler)
	router.POST("/api/files/:id/:action", routes.FileActionHandler)

	router.GET("/api/lists", routes.MediaListsHandler)
	router.POST("/api/lists", routes.MediaListsHandler)
	router.GET("/api/lists/:id", routes.MediaListHandler)
	router.PATCH("/api/lists/:id", routes.MediaListHandler)
	router.GET("/api/lists/:id/:action", routes.MediaListActionHandler)
	router.POST("/api/lists/:id/:action", routes.MediaListActionHandler)

	router.GET("/api/emails", routes.EmailsHandler)
	router.POST("/api/emails", routes.EmailsHandler)
	router.PATCH("/api/emails", routes.EmailsHandler)
	router.GET("/api/emails/:id", routes.EmailHandler)
	router.PATCH("/api/emails/:id", routes.EmailHandler)
	router.GET("/api/emails/:id/:action", routes.EmailActionHandler)

	// Security fixes
	secureMiddleware := secure.New(secure.Options{
		FrameDeny:        true,
		BrowserXssFilter: true,
	})

	// HTTP router
	app.Use(negroni.HandlerFunc(middleware.UpdateOrCreateUser))
	app.Use(negroni.HandlerFunc(middleware.AttachParameters))
	app.Use(negroni.HandlerFunc(secureMiddleware.HandlerFuncWithNext))
	app.Use(sentroni.NewRecovery(os.Getenv("SENTRY_DSN")))
	app.UseHandler(router)

	http.Handle("/", context.ClearHandler(app))
}
