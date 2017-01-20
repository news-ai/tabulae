package tasks

import (
	"fmt"
	"net/http"
)

func LetsEncryptValidation(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	fmt.Fprintf(w, "E4SjEKBbKa2JC3e2NR84axgIa2RHqzQ7pEwS_p9wsQE.DiiAVjAECHnbMqDTfs755KqRpbNq4LPmoCvd-AqUmRM")
	return
}
