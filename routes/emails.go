package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"appengine"

	"github.com/gorilla/mux"

	"github.com/news-ai/tabulae/emails"
	"github.com/news-ai/tabulae/middleware"
	"github.com/news-ai/tabulae/models"
	"github.com/news-ai/tabulae/utils"
)

func handleEmail(c appengine.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return models.GetEmail(c, id)
	case "PATCH":
		return models.UpdateSingleEmail(c, r, id)
	}
	return nil, fmt.Errorf("method not implemented")
}

func handleEmails(c appengine.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		return models.GetEmails(c, r)
	case "POST":
		return models.CreateEmail(c, r)
	case "PATCH":
		return models.UpdateBatchEmail(c, r)
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
		c.Errorf("email error: %#v", err)
		middleware.ReturnError(w, http.StatusInternalServerError, "Email handling error", err.Error())
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
			c.Errorf("email error: %#v", err)
			middleware.ReturnError(w, http.StatusInternalServerError, "Email handling error", err.Error())
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
		email, err := models.GetEmail(c, id)
		if err != nil {
			c.Errorf("email error: %#v", err)
			middleware.ReturnError(w, http.StatusInternalServerError, "Email handling error", err.Error())
			return
		}

		user, err := models.GetCurrentUser(c, r)
		if err != nil {
			c.Errorf("email error: %#v", err)
			middleware.ReturnError(w, http.StatusInternalServerError, "Email handling error", err.Error())
			return
		}

		if email.CreatedBy != user.Id {
			c.Errorf("email error: %#v", err)
			middleware.ReturnError(w, http.StatusForbidden, "Email handling error", "Not your email to send")
			return
		}

		if action == "send" {
			if email.IsSent {
				c.Errorf("email error: %#v", err)
				middleware.ReturnError(w, http.StatusInternalServerError, "Email handling error", "Email already sent")
				return
			}

			// Validate if HTML is valid
			validHTML := utils.ValidateHTML(email.Body)
			if !validHTML {
				c.Errorf("email error: %#v", err)
				middleware.ReturnError(w, http.StatusInternalServerError, "Email handling error", "Invalid HTML")
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
			c.Errorf("email error: %#v", err)
			middleware.ReturnError(w, http.StatusInternalServerError, "Email handling error", err.Error())
			return
		}

	}
}
