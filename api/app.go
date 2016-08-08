package tabulae

import (
	"net/http"
	"os"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/unrolled/secure"

	"github.com/news-ai/tabulae/auth"
	"github.com/news-ai/tabulae/middleware"
	"github.com/news-ai/tabulae/router"
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
		AllowedOrigins:   []string{"https://newsai.org", "http://localhost:3000"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		Debug:            true,
		AllowedHeaders:   []string{"*"},
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
			if len(apiRoutes[i].Routes) == 3 {
				routeName = "/{id}/{action}"
				api.HandleFunc(apiRoutes[i].HandlerName+routeName, apiRoutes[i].Routes[routeName])
			}
		}
	}

	// Setup CSRF for auth
	CSRF := csrf.Protect([]byte(os.Getenv("CSRFKEY")), csrf.Secure(false))

	// Register authentication route
	// Login with Google-based authentication
	api.HandleFunc("/auth/google", auth.GoogleLoginHandler)
	api.HandleFunc("/auth/callback", auth.GoogleCallbackHandler)

	// User registration based authentication
	api.HandleFunc("/auth/userlogin", auth.PasswordLoginHandler)
	api.HandleFunc("/auth/userregister", auth.PasswordRegisterHandler)
	api.HandleFunc("/auth", auth.PasswordLoginPageHandler)
	api.HandleFunc("/auth/registration", auth.PasswordRegisterPageHandler)
	api.HandleFunc("/auth/confirmation", auth.EmailConfirmationHandler)

	// Common
	api.HandleFunc("/auth/logout", auth.LogoutHandler)

	// Initialize the environment for a particular URL
	utils.InitURL()
	auth.SetRedirectURL()

	// Main router
	main := mux.NewRouter().StrictSlash(true)
	main.PathPrefix("/api").Handler(negroni.New(negroni.Wrap(api)))
	main.HandleFunc("/", routes.NotFoundHandler)

	// Security fixes
	secureMiddleware := secure.New(secure.Options{
		FrameDeny: true,
		// ContentSecurityPolicy: "default-src 'self'",
		BrowserXssFilter: true,
	})

	// HTTP router
	app.Use(negroni.HandlerFunc(middleware.UpdateOrCreateUser))
	app.Use(negroni.HandlerFunc(secureMiddleware.HandlerFuncWithNext))
	app.UseHandler(main)
	http.Handle("/", CSRF(context.ClearHandler(app)))
}
