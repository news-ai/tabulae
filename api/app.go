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
	"github.com/news-ai/tabulae/incoming"
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
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		Debug:            true,
		AllowedHeaders:   []string{"*"},
	})
	app.Use(c)

	// Initialize CSRF
	CSRF := csrf.Protect([]byte(os.Getenv("CSRFKEY")), csrf.Secure(false))

	r := mux.NewRouter().StrictSlash(true)

	// API router
	api := r.PathPrefix("/api").Subrouter().StrictSlash(true)

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

	// Authentication
	apiAuth := r.PathPrefix("/api/auth").Subrouter().StrictSlash(true)

	// Password based authentication
	apiAuth.Handle("/", CSRF(auth.PasswordLoginPageHandler()))
	apiAuth.Handle("/userlogin", CSRF(auth.PasswordLoginHandler()))
	apiAuth.Handle("/userregister", CSRF(auth.PasswordRegisterHandler()))
	apiAuth.Handle("/userforget", CSRF(auth.ForgetPasswordHandler()))
	apiAuth.Handle("/registration", CSRF(auth.PasswordRegisterPageHandler()))
	apiAuth.Handle("/forget", CSRF(auth.ForgetPasswordPageHandler()))
	apiAuth.Handle("/confirmation", CSRF(auth.EmailConfirmationHandler()))

	// Login with Google-based authentication
	apiAuth.HandleFunc("/google", auth.GoogleLoginHandler)
	apiAuth.HandleFunc("/callback", auth.GoogleCallbackHandler)

	// Common
	apiAuth.HandleFunc("/logout", auth.LogoutHandler)

	// Initialize the environment for a particular URL
	utils.InitURL()
	auth.SetRedirectURL()

	// Incoming from other services
	// Authentication is done through basic authentication
	apiIncoming := r.PathPrefix("/api/incoming").Subrouter().StrictSlash(true)

	apiIncoming.HandleFunc("/sendgrid", incoming.SendGridHandler)

	// Main router
	main := mux.NewRouter().StrictSlash(true)
	main.PathPrefix("/api").Handler(negroni.New(negroni.Wrap(r)))
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
	http.Handle("/", context.ClearHandler(app))
}
