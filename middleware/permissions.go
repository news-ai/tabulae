package middleware

import (
	e "github.com/news-ai/jsonerror"
	"gopkg.in/unrolled/render.v1"
	"net/http"
)

type Error struct {
	ErrorMessage error `json:"error"`
	Message      error `json:"message"`
	StatusCode   int   `json:"code"`
}

func errorBase() map[string][]map[string]string {
	errors := []map[string]string{}
	errorBase := map[string][]map[string]string{}
	errorBase["errors"] = errors
	return errorBase
}

func ReturnError(w http.ResponseWriter, errorCode int, messageOne string, messageTwo string) {
	errors := errorBase()
	err := e.New(errorCode, messageOne, messageTwo)
	r := render.New(render.Options{})
	errors["errors"] = append(errors["errors"], err.Render())
	r.JSON(w, errorCode, errors)
	return
}
