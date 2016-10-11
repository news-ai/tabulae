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

	importedFiles, err := controllers.FilterFileByImported(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		errors.ReturnError(w, http.StatusInternalServerError, "Could not get files", err.Error())
		return
	}

	log.Infof(c, "%v", importedFiles)

	for i := 0; i < len(importedFiles); i++ {
		log.Infof(c, "%v", importedFiles[i].FileName)
		err = files.DeleteFile(r, importedFiles[i].FileName)
		if err != nil {
			log.Errorf(c, "%v", err)
		} else {
			importedFiles[i].FileExists = false
			importedFiles[i].Save(c)
		}
	}

	// If successful
	w.WriteHeader(200)
	return
}
