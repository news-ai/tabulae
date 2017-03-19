package routes

func handleDatabases(c context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		val, included, count, err := controllers.GetDatabases(c, r)
		return api.BaseResponseHandler(val, included, count, err, r)
	}
	return nil, errors.New("method not implemented")
}

// Handler for when the user wants all the contacts.
func DatabasesHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	val, err := handleDatabases(c, w, r)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Database handling error", err.Error())
	}
	return
}
