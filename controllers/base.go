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
	order = strings.ToLower(order)
	order = strings.Title(order)

	if normalizedOrder, ok := normalized[order]; ok {
		return normalizedOrder
	}

	return order
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
