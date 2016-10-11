package tasks

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/files"

	"github.com/news-ai/web/errors"
)

func RemoveImportedFilesHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	files, err := controllers.FilterFileByImported(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		errors.ReturnError(w, http.StatusInternalServerError, "Could not get files", "Problem getting all imported files")
		return
	}

	for i := 0; i < len(files); i++ {
		err = files.DeleteFile(r, files[i].FileName)
		if err != nil {
			log.Errorf(c, "%v", err)
			errors.ReturnError(w, http.StatusInternalServerError, "Could not delete files", err.Error())
			return
		}
		files[i].FileExists = false
		files[i].Save(c)
	}

	// If successful
	w.WriteHeader(200)
	return
}
