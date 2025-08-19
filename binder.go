package teta

import (
	json "encoding/json/v2"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type defaultBinder struct {
	r *http.Request
}

// type Binder interface {
// 	Bind(dest any) error
// 	BindQuery(dest any) error
// 	BindBody(dest any) error
// 	BindPath(dest any) error
// }

func (b *defaultBinder) bind(r *http.Request, dest any) error {
	if err := b.bindQuery(r, dest); err != nil {
		return fmt.Errorf("query bind failed: %w", err)
	}

	if err := b.bindPath(r, dest); err != nil {
		return fmt.Errorf("path bind failed: %w", err)
	}

	if hasJSONTags(dest) {
		if err := b.bindBody(r, dest); err != nil {
			return fmt.Errorf("body bind failed: %w", err)
		}
	}

	return nil
}

func (b *defaultBinder) bindQuery(r *http.Request, dest any) error {
	rv, err := checkIfPoiner(dest)
	if err != nil {
		return fmt.Errorf("dest must be a pointer to a struct")
	}

	rt := rv.Type()
	queryValues := r.URL.Query()

	for i := range rt.NumField() {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		queryTag := field.Tag.Get("query")
		if queryTag == "" {
			continue
		}

		queryValue := queryValues.Get(queryTag)
		if queryValue == "" {
			continue
		}

		if err := setFieldFromString(fieldValue, queryValue); err != nil {
			return fmt.Errorf("field %s: %w", field.Name, err)
		}
	}

	return nil
}

// Аналогично обновляем bindPath, bindBody
func (b *defaultBinder) bindPath(r *http.Request, dest any) error {
	rv, err := checkIfPoiner(dest)
	if err != nil {
		return fmt.Errorf("dest must be a pointer to a struct")
	}

	rt := rv.Type()

	for i := range rt.NumField() {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		pathTag := field.Tag.Get("path")
		if pathTag == "" {
			continue
		}

		pathValue := r.PathValue(pathTag)
		if pathValue == "" {
			continue
		}

		if err := setFieldFromString(fieldValue, pathValue); err != nil {
			return fmt.Errorf("field %s: %w", field.Name, err)
		}
	}

	return nil
}

func (b *defaultBinder) bindBody(r *http.Request, dest any) error {
	contentType := r.Header.Get("Content-Type")

	switch {
	case strings.Contains(contentType, "application/json"):
		if err := json.UnmarshalRead(r.Body, &dest); err != nil {
			return fmt.Errorf("json decode failed: %w", err)
		}

	case strings.Contains(contentType, "application/x-www-form-urlencoded"),
		strings.Contains(contentType, "multipart/form-data"):
		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("form parse failed: %w", err)
		}
	default:
		return fmt.Errorf("unsupported content-type: %s", contentType)
	}

	return nil
}

func setFieldFromString(field reflect.Value, value string) error {
	if !field.CanSet() {
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)
	default:
		return fmt.Errorf("unsupported type: %s", field.Kind())
	}

	return nil
}

// checkIfPoiner returns reflect.Value.Elem() if dest is pointer othervise returns dest and error
func checkIfPoiner(dest any) (reflect.Value, error) {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Struct {
		return rv, fmt.Errorf("dest must be a pointer to a struct")
	}

	return rv.Elem(), nil
}

func hasJSONTags(dest any) bool {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Struct {
		return false
	}

	rt := rv.Elem().Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			return true
		}
	}
	return false
}
