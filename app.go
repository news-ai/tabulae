package tabulae

import (
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/user"
)

func init() {
	// Register the index handler to the
	// default DefaultServeMux.
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/user", userIndex)
}

func handleIndex(res http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)
	u := user.Current(ctx)
	url, _ := user.LogoutURL(ctx, "/")
	res.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(res, `Welcome, %s! (<a href="%s">sign out</a>)`, u, url)
}

func userIndex(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	ctx := appengine.NewContext(req)
	u := user.Current(ctx)

	encoder := json.NewEncoder(res)
	encoder.Encode(u)
}
