package utils

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"

	"github.com/labstack/echo"
)

//MissingFields verify if fields are missing
func MissingFields(c echo.Context, val reflect.Value, keys []string) error {
	var missingFields []string

	for _, key := range keys {
		if val.FieldByName(key).String() == "" {
			missingFields = append(missingFields, strings.ToLower(key))
		}
	}
	if len(missingFields) > 0 {
		var err error
		err = errors.New("Some fields missing: " + strings.Join(missingFields, ", "))
		return err
	}
	return nil
}

//NewNullString transform string into nullstring
func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}
