package utils

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/sinha-shaurya/agentic-bank/utils/httperr"
)

// validate is built once: the instance caches struct metadata by type, which a
// per-request constructor throws away. Messages name the json field, not the
// Go field.
var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})
	return v
}

// BindAndValidate fills obj from the URI params, query string and JSON body,
// then runs its `validate` tags. Every failure returns an httperr.HttpError
// carrying 400 — recover it with errors.As.
func BindAndValidate(ctx *gin.Context, obj any) error {
	if err := ctx.ShouldBindUri(obj); err != nil {
		return badRequest(err)
	}

	if err := ctx.ShouldBindQuery(obj); err != nil {
		return badRequest(err)
	}

	if hasRequestBody(ctx.Request) {
		if err := ctx.ShouldBindJSON(obj); err != nil {
			return badRequest(err)
		}
	}

	if err := validate.Struct(obj); err != nil {
		var fieldErrs validator.ValidationErrors
		if errors.As(err, &fieldErrs) {
			messages := make([]string, 0, len(fieldErrs))
			for _, fieldErr := range fieldErrs {
				messages = append(messages, validationMessage(fieldErr))
			}
			return badRequest(errors.New(strings.Join(messages, "; ")))
		}
		return badRequest(err)
	}

	return nil
}

func badRequest(err error) error {
	return httperr.NewCode(httperr.ErrValidation, err)
}

// hasRequestBody gates JSON binding on the request, not on the DTO. Keying off
// "the struct has a json tag" decodes an empty GET body and answers a valid
// GET /wallets/{id} with a 400.
func hasRequestBody(req *http.Request) bool {
	if req == nil || req.Body == nil || req.Body == http.NoBody {
		return false
	}

	switch req.Method {
	case http.MethodGet, http.MethodHead, http.MethodDelete, http.MethodOptions:
		return false
	}

	// -1 is chunked with an unknown length, which is a body.
	return req.ContentLength != 0
}

// validationMessage renders one field error. A failing `gt=0` on amount_paise
// must not be reported as "is required".
//nolint
func validationMessage(fieldErr validator.FieldError) string {
	field := fieldErr.Field()
	param := fieldErr.Param()

	switch fieldErr.Tag() {
	case "required":
		return field + " is required"
	case "gt":
		return field + " must be greater than " + param
	case "gte", "min":
		return field + " must be at least " + param
	case "lt":
		return field + " must be less than " + param
	case "lte", "max":
		return field + " must be at most " + param
	case "len":
		return field + " must be exactly " + param + " in length"
	case "oneof":
		return field + " must be one of: " + strings.Join(strings.Fields(param), ", ")
	case "uuid", "uuid4":
		return field + " must be a valid UUID"
	case "email":
		return field + " must be a valid email address"
	case "nefield", "necsfield":
		return field + " must not equal " + param
	default:
		return field + " is invalid (" + fieldErr.Tag() + ")"
	}
}
