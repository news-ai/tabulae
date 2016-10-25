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

	gaeTasks "github.com/news-ai/gaesessions/tasks"

	"github.com/news-ai/tabulae/auth"
	"github.com/news-ai/tabulae/incoming"
	"github.com/news-ai/tabulae/middleware"
	"github.com/news-ai/tabulae/notifications"
	"github.com/news-ai/tabulae/routes"
	"github.com/news-ai/tabulae/search"
	"github.com/news-ai/tabulae/tasks"
	"github.com/news-ai/tabulae/utils"

	"github.com/news-ai/web/api"
	commonMiddleware "github.com/news-ai/web/middleware"
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
	search.InitializeElasticSearch()

	// Initialize router
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(api.NotFound)

	// Not found Handler
	router.GET("/", api.NotFoundHandler)
	router.GET("/api", api.NotFoundHandler)

	/*
	* Authentication Handler
	 */

	// Login
	router.Handler("GET", "/api/auth", CSRF(auth.PasswordLoginPageHandler()))
	router.Handler("POST", "/api/auth/userlogin", CSRF(auth.PasswordLoginHandler()))

	// Forget password
	router.Handler("GET", "/api/auth/forget", CSRF(auth.ForgetPasswordPageHandler()))
	router.Handler("POST", "/api/auth/userforget", CSRF(auth.ForgetPasswordHandler()))

	// Change password
	router.Handler("GET", "/api/auth/changepassword", CSRF(auth.ChangePasswordPageHandler()))
	router.Handler("POST", "/api/auth/userchange", CSRF(auth.ChangePasswordHandler()))

	// Reset password
	router.Handler("GET", "/api/auth/resetpassword", CSRF(auth.ResetPasswordPageHandler()))
	router.Handler("POST", "/api/auth/userreset", CSRF(auth.ResetPasswordHandler()))

	// Register user
	router.Handler("GET", "/api/auth/registration", CSRF(auth.PasswordRegisterPageHandler()))
	router.Handler("POST", "/api/auth/userregister", CSRF(auth.PasswordRegisterHandler()))

	// Email confirmation
	router.Handler("GET", "/api/auth/confirmation", CSRF(auth.EmailConfirmationHandler()))

	// Login with Google
	router.GET("/api/auth/google", auth.GoogleLoginHandler)
	router.GET("/api/auth/googlecallback", auth.GoogleCallbackHandler)

	// Internal auth: Linkedin
	router.GET("/api/internal_auth/linkedin", auth.LinkedinLoginHandler)
	router.GET("/api/internal_auth/linkedincallback", auth.LinkedinCallbackHandler)

	// Internal auth: Instagram
	router.GET("/api/internal_auth/instagram", auth.InstagramLoginHandler)
	router.GET("/api/internal_auth/instagramcallback", auth.InstagramCallbackHandler)

	// Logout user
	router.GET("/api/auth/logout", auth.LogoutHandler)

	/*
	* Billing Handler
	 */

	// Start a free trial
	router.Handler("GET", "/api/billing/plans/trial", CSRF(auth.TrialPlanPageHandler()))
	router.Handler("POST", "/api/billing/plans/start-trial", CSRF(auth.ChooseTrialPlanHandler()))

	// Get all the plans
	router.Handler("GET", "/api/billing/plans", CSRF(auth.ChoosePlanPageHandler()))

	// Add payment method
	router.Handler("GET", "/api/billing/payment-methods", CSRF(auth.PaymentMethodsPageHandler()))
	router.Handler("POST", "/api/billing/add-payment-method", CSRF(auth.PaymentMethodsHandler()))

	// Main billing page for a user
	router.Handler("GET", "/api/billing", CSRF(auth.BillingPageHandler()))

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
	router.GET("/api/users/:id/:action", routes.UserActionHandler)

	router.GET("/api/agencies", routes.AgenciesHandler)
	router.GET("/api/agencies/:id", routes.AgencyHandler)

	router.GET("/api/publications", routes.PublicationsHandler)
	router.POST("/api/publications", routes.PublicationsHandler)
	router.GET("/api/publications/:id", routes.PublicationHandler)
	router.GET("/api/publications/:id/:action", routes.PublicationActionHandler)

	router.GET("/api/contacts", routes.ContactsHandler)
	router.POST("/api/contacts", routes.ContactsHandler)
	router.PATCH("/api/contacts", routes.ContactsHandler)
	router.GET("/api/contacts/:id", routes.ContactHandler)
	router.PATCH("/api/contacts/:id", routes.ContactHandler)
	router.DELETE("/api/contacts/:id", routes.ContactHandler)
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

	router.GET("/api/templates", routes.TemplatesHandler)
	router.POST("/api/templates", routes.TemplatesHandler)
	router.GET("/api/templates/:id", routes.TemplateHandler)
	router.PATCH("/api/templates/:id", routes.TemplateHandler)

	router.GET("/api/feeds", routes.FeedsHandler)
	router.POST("/api/feeds", routes.FeedsHandler)
	router.GET("/api/feeds/:id", routes.FeedHandler)
	router.DELETE("/api/feeds/:id", routes.FeedHandler)

	router.GET("/api/notifications", routes.NotificationsHandler)

	router.GET("/api/invites", routes.InvitesHandler)
	router.POST("/api/invites", routes.InvitesHandler)

	// Security fixes
	secureMiddleware := secure.New(secure.Options{
		FrameDeny:        true,
		BrowserXssFilter: true,
	})

	// HTTP router
	app.Use(negroni.HandlerFunc(commonMiddleware.AppEngineCheck))
	app.Use(negroni.HandlerFunc(middleware.UpdateOrCreateUser))
	app.Use(negroni.HandlerFunc(commonMiddleware.AttachParameters))
	app.Use(negroni.HandlerFunc(secureMiddleware.HandlerFuncWithNext))
	app.Use(sentroni.NewRecovery(os.Getenv("SENTRY_DSN")))
	app.UseHandler(router)

	/*
	* Tasks Handler
	 */

	// Tasks needing to have middleware
	router.POST("/tasks/socialUsernameInvalid", tasks.SocialUsernameInvalid)
	router.POST("/tasks/feedInvalid", tasks.FeedInvalid)

	// Tasks needing to not have middleware
	http.HandleFunc("/tasks/removeExpiredSessions", gaeTasks.RemoveExpiredSessionsHandler)
	http.HandleFunc("/tasks/removeImportedFiles", tasks.RemoveImportedFilesHandler)

	/*
	* Appengine Handler
	 */

	http.HandleFunc("/_ah/channel/connected/", notifications.UserConnect)
	http.HandleFunc("/_ah/channel/disconnected/", notifications.UserDisconnect)

	// Register the app router
	http.Handle("/", context.ClearHandler(app))
}
