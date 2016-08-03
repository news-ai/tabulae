package models

import (
	"time"
)

type CustomContactField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Contact struct {
	Id int64 `json:"id" datastore:"-"`

	// Contact information
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`

	// Notes on a particular contact
	Notes string `json:"notes"`

	// Publications this contact works for now and before
	Employers     []int64 `json:"employers"`
	PastEmployers []int64 `json:"pastemployers"`

	// Social information
	LinkedIn  string `json:"linkedin"`
	Twitter   string `json:"twitter"`
	Instagram string `json:"instagram"`
	MuckRack  string `json:"-"`
	Website   string `json:"website"`
	Blog      string `json:"blog"`

	// Custom fields
	CustomFields []CustomContactField `json:"customfields"`

	// Parent contact
	IsMasterContact bool  `json:"ismastercontact"`
	ParentContact   int64 `json:"parent"`

	// Is information outdated
	IsOutdated bool `json:"isoutdated"`

	CreatedBy int64 `json:"createdby"`

	Created         time.Time `json:"created"`
	Updated         time.Time `json:"updated"`
	LinkedInUpdated time.Time `json:"linkedinupdated"`
}
