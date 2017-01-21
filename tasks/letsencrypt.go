package tasks

import (
	"fmt"
	"net/http"
)

func LetsEncryptValidation(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	fmt.Fprintf(w, "cK4fsCUnbTkr0j5fkuYM-V9NwoET2JP_4iRTBBvLDUE.oKzw5QYkN1q8zhhAe-jdS_VxeEiIqz4MC7pSvnuwGq4")
	return
}
