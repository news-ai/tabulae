package tasks

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/controllers"

	"github.com/news-ai/web/errors"
)

func RemoveImportedFilesHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	files, err := controllers.FilterFileByImported(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		errors.ReturnError(w, http.StatusInternalServerError, "Could not get files", err)
		return
	}

	for i := 0; i < len(files); i++ {
		log.Infof(c, "%v", files[i].FileName)
	}

	// If successful
	w.WriteHeader(200)
	return
}
