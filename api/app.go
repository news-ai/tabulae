package tabulae

import (
	"net/http"
	"os"

	"github.com/bradleyg/go-sentroni"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
	"github.com/gorilla/csrf"
	// "github.com/gorilla/mux"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"github.com/unrolled/secure"

	"github.com/news-ai/tabulae/auth"
	"github.com/news-ai/tabulae/incoming"
	"github.com/news-ai/tabulae/middleware"
	// "github.com/news-ai/tabulae/router"
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

	// // Initialize the environment for a particular URL
	utils.InitURL()
	auth.SetRedirectURL()

	router := httprouter.New()

	router.GET("/", routes.NotFoundHandler)
	router.GET("/api", routes.NotFoundHandler)

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

	router.POST("/api/incoming/sendgrid", incoming.SendGridHandler)

	// // Register routes
	// apiRoutes := router.GetRoutes()
	// for i := 0; i < len(apiRoutes); i++ {
	// 	api.HandleFunc(apiRoutes[i].HandlerName, apiRoutes[i].Routes["/"])
	// 	if len(apiRoutes[i].Routes) > 1 {
	// 		routeName := "/{id}"
	// 		api.HandleFunc(apiRoutes[i].HandlerName+routeName, apiRoutes[i].Routes[routeName])
	// 		if len(apiRoutes[i].Routes) == 3 {
	// 			routeName = "/{id}/{action}"
	// 			api.HandleFunc(apiRoutes[i].HandlerName+routeName, apiRoutes[i].Routes[routeName])
	// 		}
	// 	}
	// }

	// // Main router
	// main := mux.NewRouter().StrictSlash(true)
	// main.PathPrefix("/api").Handler(negroni.New(negroni.Wrap(r)))
	// main.HandleFunc("/", routes.NotFoundHandler)

	// Security fixes
	secureMiddleware := secure.New(secure.Options{
		FrameDeny: true,
		// ContentSecurityPolicy: "default-src 'self'",
		BrowserXssFilter: true,
	})

	// // Setup error logging
	dsn := "https://eccfde5212974b1b9c284995885d0446:592890e0beac49a58d410f1e7061983e@app.getsentry.com/92935"

	// HTTP router
	app.Use(negroni.HandlerFunc(middleware.UpdateOrCreateUser))
	app.Use(negroni.HandlerFunc(secureMiddleware.HandlerFuncWithNext))
	app.Use(sentroni.NewRecovery(dsn))
	app.UseHandler(router)

	http.Handle("/", context.ClearHandler(app))
}
