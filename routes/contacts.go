package routes

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"

	"github.com/julienschmidt/httprouter"
	"github.com/pquerna/ffjson/ffjson"

	"github.com/news-ai/tabulae/controllers"

	"github.com/news-ai/web/api"
	nError "github.com/news-ai/web/errors"
)

func handleContactAction(c context.Context, r *http.Request, id string, action string) (interface{}, error) {
	switch r.Method {
	case "GET":
		switch action {
		case "feed":
			val, included, count, total, err := controllers.GetFeedForContact(c, r, id)
			return api.BaseResponseHandler(val, included, count, total, err, r)
		case "headlines":
			val, included, count, total, err := controllers.GetHeadlinesForContact(c, r, id)
			return api.BaseResponseHandler(val, included, count, total, err, r)
		case "tweets":
			val, included, count, total, err := controllers.GetTweetsForContact(c, r, id)
			return api.BaseResponseHandler(val, included, count, total, err, r)
		case "twitterprofile":
			return api.BaseSingleResponseHandler(controllers.GetTwitterProfileForContact(c, r, id))
		case "twittertimeseries":
			return api.BaseSingleResponseHandler(controllers.GetTwitterTimeseriesForContact(c, r, id))
		case "instagrams":
			val, included, count, total, err := controllers.GetInstagramPostsForContact(c, r, id)
			return api.BaseResponseHandler(val, included, count, total, err, r)
		case "instagramprofile":
			return api.BaseSingleResponseHandler(controllers.GetInstagramProfileForContact(c, r, id))
		case "instagramtimeseries":
			return api.BaseSingleResponseHandler(controllers.GetInstagramTimeseriesForContact(c, r, id))
		case "feeds":
			val, included, count, total, err := controllers.GetFeedsForContact(c, r, id)
			return api.BaseResponseHandler(val, included, count, total, err, r)
		case "emails":
			val, included, count, total, err := controllers.GetEmailsForContact(c, r, id)
			return api.BaseResponseHandler(val, included, count, total, err, r)
		case "lists":
			val, included, count, total, err := controllers.GetListsForContact(c, r, id)
			return api.BaseResponseHandler(val, included, count, total, err, r)
		case "enrich":
			return api.BaseSingleResponseHandler(controllers.EnrichContact(c, r, id))
		}
	}
	return nil, errors.New("method not implemented")
}

func handleContact(c context.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return api.BaseSingleResponseHandler(controllers.GetContact(c, r, id))
	case "PATCH":
		return api.BaseSingleResponseHandler(controllers.UpdateSingleContact(c, r, id))
	case "POST":
		if id == "copy" {
			val, included, count, total, err := controllers.CopyContacts(c, r)
			return api.BaseResponseHandler(val, included, count, total, err, r)
		} else if id == "bulkdelete" {
			val, included, count, total, err := controllers.BulkDeleteContacts(c, r)
			return api.BaseResponseHandler(val, included, count, total, err, r)
		}
	case "DELETE":
		return api.BaseSingleResponseHandler(controllers.DeleteContact(c, r, id))
	}
	return nil, errors.New("method not implemented")
}

func handleContacts(c context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		val, included, count, total, err := controllers.GetContacts(c, r)
		return api.BaseResponseHandler(val, included, count, total, err, r)
	case "POST":
		val, included, count, total, err := controllers.CreateContact(c, r)
		return api.BaseResponseHandler(val, included, count, total, err, r)
	case "PATCH":
		val, included, count, total, err := controllers.UpdateBatchContact(c, r)
		return api.BaseResponseHandler(val, included, count, total, err, r)
	}
	return nil, errors.New("method not implemented")
}

// Handler for when the user wants all the contacts.
func ContactsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	val, err := handleContacts(c, w, r)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Contact handling error", err.Error())
	}
	return
}

// Handler for when there is a key present after /users/<id> route.
func ContactHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	id := ps.ByName("id")
	val, err := handleContact(c, r, id)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Contact handling error", err.Error())
	}
	return
}

func ContactActionHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	id := ps.ByName("id")
	action := ps.ByName("action")
	val, err := handleContactAction(c, r, id, action)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Contact handling error", err.Error())
	}
	return
}
