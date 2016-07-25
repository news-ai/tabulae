package routes

import (
	"net/http"

	"google.golang.org/appengine"

	"github.com/news-ai/tabulae/middleware"

	"google.golang.org/cloud/storage"
)

// Handler for when the user wants all the users.
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		w.Header().Set("Content-Type", "application/json")
		c := appengine.NewContext(r)

		cleanUp := []string{}

		fileName := "hello.xlsx"

		bucket, err := getStorageBucket(r, "")
		if err != nil {
			middleware.ReturnError(w, http.StatusInternalServerError, "Upload handling error", err.Error())
			return
		}

		client, err := storage.NewClient(c)
		if err != nil {
			middleware.ReturnError(w, http.StatusInternalServerError, "Upload handling error", err.Error())
			return
		}
		defer client.Close()

		clientBucket := client.Bucket(bucket)

		wc := clientBucket.Object(fileName).NewWriter(c)
		wc.ContentType = "text/plain"
		wc.Metadata = map[string]string{
			"x-goog-meta-foo": "foo",
			"x-goog-meta-bar": "bar",
		}
		cleanUp = append(cleanUp, fileName)

		if _, err := wc.Write([]byte("abcde\n")); err != nil {
			middleware.ReturnError(w, http.StatusInternalServerError, "Upload handling error", err.Error())
			return
		}
		if err := wc.Close(); err != nil {
			middleware.ReturnError(w, http.StatusInternalServerError, "Upload handling error", err.Error())
			return
		}

		middleware.ReturnError(w, http.StatusInternalServerError, "Upload handling error", "")
		return
	}
	middleware.ReturnError(w, http.StatusInternalServerError, "Upload handling error", "Method not implemented")
}
