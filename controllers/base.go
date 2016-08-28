package controllers

import (
	"net/http"
	"strings"

	"google.golang.org/appengine/datastore"

	gcontext "github.com/gorilla/context"
)

// At some point automate this
var normalized = map[string]string{
	"createdby":          "CreatedBy",
	"firstname":          "FirstName",
	"lastname":           "LastName",
	"pastemployers":      "PastEmployers",
	"muckrack":           "MuckRack",
	"customfields":       "CustomFields",
	"ismastercontact":    "IsMasterContact",
	"parentcontact":      "ParentContact",
	"isoutdated":         "IsOutdated",
	"linkedinupdated":    "LinkedInUpdated",
	"sendgridid":         "SendGridId",
	"bouncedreason":      "BouncedReason",
	"issent":             "IsSent",
	"filename":           "FileName",
	"listid":             "ListId",
	"fileexists":         "FileExists",
	"fieldsmap":          "FieldsMap",
	"noticationid":       "NoticationId",
	"objectid":           "ObjectId",
	"noticationobjectid": "NoticationObjectId",
	"canwrite":           "CanWrite",
	"userid":             "UserId",
	"googleid":           "GoogleId",
	"apikey":             "ApiKey",
	"emailconfirmed":     "EmailConfirmed",
}

func normalizeOrderQuery(order string) string {
	operator := ""
	if string(order[0]) == "-" {
		operator = string(order[0])
		order = order[1:]
	}

	order = strings.ToLower(order)

	// If it is inside the abnormal cases above
	if normalizedOrder, ok := normalized[order]; ok {
		return operator + normalizedOrder
	}

	// Else return the titled version of it
	order = strings.Title(order)
	return operator + order
}

func constructQuery(query *datastore.Query, r *http.Request) *datastore.Query {
	order := gcontext.Get(r, "order").(string)
	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	if order != "" {
		query = query.Order(normalizeOrderQuery(order))
	}

	return query.Limit(limit).Offset(offset)
}
