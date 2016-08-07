package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"

	"github.com/gorilla/mux"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/emails"
	"github.com/news-ai/tabulae/permissions"
	"github.com/news-ai/tabulae/utils"
)

func handleEmail(c context.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return controllers.GetEmail(c, r, id)
	case "PATCH":
		return controllers.UpdateSingleEmail(c, r, id)
	}
	return nil, fmt.Errorf("method not implemented")
}

func handleEmails(c context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		return controllers.GetEmails(c, r)
	case "POST":
		return controllers.CreateEmail(c, r)
	case "PATCH":
		return controllers.UpdateBatchEmail(c, r)
	}
	return nil, fmt.Errorf("method not implemented")
}

// Handler for when the user wants all the contacts.
func EmailsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	val, err := handleEmails(c, w, r)

	if err == nil {
		err = json.NewEncoder(w).Encode(val)
	}

	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "Email handling error", err.Error())
		return
	}
}

// Handler for when there is a key present after /users/<id> route.
func EmailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	// If there is an ID
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if ok {
		val, err := handleEmail(c, r, id)

		if err == nil {
			err = json.NewEncoder(w).Encode(val)
		}

		if err != nil {
			permissions.ReturnError(w, http.StatusInternalServerError, "Email handling error", err.Error())
			return
		}
	}
}

func EmailActionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	// If there is an ID
	vars := mux.Vars(r)
	id, idOk := vars["id"]
	action, actionOk := vars["action"]
	if idOk && actionOk {
		email, err := controllers.GetEmail(c, r, id)
		if err != nil {
			permissions.ReturnError(w, http.StatusInternalServerError, "Email handling error", err.Error())
			return
		}

		user, err := controllers.GetCurrentUser(c, r)
		if err != nil {
			permissions.ReturnError(w, http.StatusInternalServerError, "Email handling error", err.Error())
			return
		}

		if email.CreatedBy != user.Id {
			permissions.ReturnError(w, http.StatusForbidden, "Email handling error", "Not your email to send")
			return
		}

		if action == "send" {
			if email.IsSent {
				permissions.ReturnError(w, http.StatusInternalServerError, "Email handling error", "Email already sent")
				return
			}

			// Validate if HTML is valid
			validHTML := utils.ValidateHTML(email.Body)
			if !validHTML {
				permissions.ReturnError(w, http.StatusInternalServerError, "Email handling error", "Invalid HTML")
				return
			}

			emailSent := emails.SendEmail(r, email, user)
			if emailSent {
				val, err := email.MarkSent(c)
				if err == nil {
					err = json.NewEncoder(w).Encode(val)
				}
			}
		}

		if err != nil {
			permissions.ReturnError(w, http.StatusInternalServerError, "Email handling error", err.Error())
			return
		}

	}
}
