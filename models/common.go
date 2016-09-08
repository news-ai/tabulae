package models

import (
	"errors"
	"reflect"

	"github.com/news-ai/cast"
)

func SetField(obj interface{}, name string, value interface{}) error {
	if name == "id" {
		name = "Id"
	}

	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		return errors.New("No such field:" + name + " in obj")
	}

	if !structFieldValue.CanSet() {
		return errors.New("Cannot set" + name + " field value")
	}

	// Cast time
	if name == "Created" || name == "Updated" || name == "LinkedInUpdated" {
		returnValue := cast.ToTime(value)
		val := reflect.ValueOf(returnValue)
		structFieldValue.Set(val)
		return nil
	}

	// CustomFields
	if name == "CustomFields" {
		val := reflect.ValueOf(value)
		structFieldValue.Set(val)
		return nil
	}

	// Int64
	if name == "Id" || name == "CreatedBy" || name == "ParentContact" || name == "ListId" {
		returnValue := cast.ToInt64(value)
		val := reflect.ValueOf(returnValue)
		structFieldValue.Set(val)
		return nil
	}

	// Int64 array
	if name == "Administrators" || name == "Employers" || name == "PastEmployers" {
		returnValue, err := cast.ToInt64SliceE(value)
		if err != nil {
			return err
		}
		val := reflect.ValueOf(returnValue)
		structFieldValue.Set(val)
		return nil
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)

	if structFieldType != val.Type() {
		invalidTypeError := errors.New("Provided value type didn't match obj field type")
		return invalidTypeError
	}

	structFieldValue.Set(val)
	return nil
}
