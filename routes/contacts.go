package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"golang.org/x/net/context"

	"google.golang.org/appengine"

	"github.com/gorilla/mux"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/permissions"
)

func handleContact(c context.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return controllers.GetContact(c, r, id)
	case "PATCH":
		return controllers.UpdateSingleContact(c, r, id)
	}
	return nil, fmt.Errorf("method not implemented")
}

func handleContacts(c context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		return controllers.GetContacts(c, r)
	case "POST":
		return controllers.CreateContact(c, r)
	case "PATCH":
		return controllers.UpdateBatchContact(c, r)
	}
	return nil, fmt.Errorf("method not implemented")
}

// Handler for when the user wants all the contacts.
func ContactsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	val, err := handleContacts(c, w, r)

	if err == nil {
		err = json.NewEncoder(w).Encode(val)
	}

	if err != nil {
		permissions.ReturnError(w, http.StatusInternalServerError, "Contact handling error", err.Error())
		return
	}
}

// Handler for when there is a key present after /users/<id> route.
func ContactHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	// If there is an ID
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if ok {
		val, err := handleContact(c, r, id)

		if err == nil {
			err = json.NewEncoder(w).Encode(val)
		}

		if err != nil {
			permissions.ReturnError(w, http.StatusInternalServerError, "Contact handling error", err.Error())
			return
		}
	}
}

func ContactActionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)

	// If there is an ID
	vars := mux.Vars(r)
	id, ok := vars["id"]
	action, actionOk := vars["action"]
	if ok && actionOk {
		// Get current contact
		contact, err := controllers.GetContact(c, r, id)
		if err != nil {
			permissions.ReturnError(w, http.StatusInternalServerError, "Contact handling error", err.Error())
			return
		}

		// Get parent contact
		parentContact, err := controllers.GetContact(c, r, strconv.FormatInt(contact.ParentContact, 10))
		if err != nil {
			permissions.ReturnError(w, http.StatusInternalServerError, "Contact handling error", err.Error())
			return
		}

		// Two actions: diff, update
		if action == "diff" {
			newEmployers := []string{}
			for i := 0; i < len(parentContact.Employers); i++ {
				// Get each publication
				currentPublication, err := controllers.GetPublication(c, strconv.FormatInt(parentContact.Employers[i], 10))
				if err != nil {
					permissions.ReturnError(w, http.StatusInternalServerError, "Contact handling error", "Only actions are diff and update")
					return
				}
				newEmployers = append(newEmployers, currentPublication.Name)
			}
			data := struct {
				Changes []string `json:"changes"`
			}{
				newEmployers,
			}

			json.NewEncoder(w).Encode(data)
			return
		} else if action == "update" {
			if !contact.IsOutdated {
				err = json.NewEncoder(w).Encode(contact)
				return
			}

			val, err := controllers.UpdateContactToParent(c, r, &contact)

			if err == nil {
				err = json.NewEncoder(w).Encode(val)
				return
			}

			if err != nil {
				permissions.ReturnError(w, http.StatusInternalServerError, "Contact handling error", err.Error())
				return
			}
		}

		permissions.ReturnError(w, http.StatusInternalServerError, "Contact handling error", "Only actions are diff and update")
		return
	}
}
