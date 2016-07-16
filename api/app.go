package tabulae

import (
	"fmt"
	"net/http"

	"appengine"
	"appengine/user"

	"github.com/news-ai/tabulae/routes"
)

func init() {
	// Register the index handler to the
	// default DefaultServeMux.
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/user", routes.UserHandler)
}

func handleIndex(res http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)
	u := user.Current(ctx)
	url, _ := user.LogoutURL(ctx, "/")
	res.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(res, `Welcome, %s! (<a href="%s">sign out</a>)`, u, url)
}
