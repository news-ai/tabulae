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

	for i := 0; i < len(importedFiles); i++ {
		err = files.DeleteFile(r, importedFiles[i].FileName)
		if err != nil {
			if err.Error() == "storage: object doesn't exist" {
				importedFiles[i].FileExists = false
				importedFiles[i].Save(c)
			} else {
				log.Errorf(c, "%v", importedFiles[i].FileName)
				log.Errorf(c, "%v", err)
			}
		} else {
			importedFiles[i].FileExists = false
			importedFiles[i].Save(c)
		}
	}

	// If successful
	w.WriteHeader(200)
	return
}
