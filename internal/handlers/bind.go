package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func bindErrorMessage(err error, req any) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		msgs := make([]string, 0, len(ve))
		for _, fe := range ve {
			msgs = append(msgs, jsonName(req, fe.StructField())+" "+ruleMessage(fe))
		}
		return strings.Join(msgs, "; ")
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return fmt.Sprintf("%s must be of type %s", typeErr.Field, typeErr.Type)
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) || errors.Is(err, io.ErrUnexpectedEOF) {
		return "request body is not valid JSON"
	}
	if errors.Is(err, io.EOF) {
		return "request body is empty"
	}
	if field, ok := strings.CutPrefix(err.Error(), "json: unknown field "); ok {
		return fmt.Sprintf("unknown field %s", field)
	}

	return "invalid request body"
}

func jsonName(req any, field string) string {
	t := reflect.TypeOf(req)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if f, ok := t.FieldByName(field); ok {
		if name, _, _ := strings.Cut(f.Tag.Get("json"), ","); name != "" && name != "-" {
			return name
		}
	}
	return field
}

func ruleMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		switch fe.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return "is required and must be greater than 0"
		}
		return "is required"
	case "min":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("must be at least %s characters long", fe.Param())
		}
		return fmt.Sprintf("must be at least %s", fe.Param())
	case "max":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("must be at most %s characters long", fe.Param())
		}
		return fmt.Sprintf("must be at most %s", fe.Param())
	case "lte":
		return fmt.Sprintf("must be at most %s", fe.Param())
	case "gt":
		return fmt.Sprintf("must be greater than %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", strings.ReplaceAll(fe.Param(), " ", ", "))
	default:
		return fmt.Sprintf("failed the %s rule", fe.Tag())
	}
}
