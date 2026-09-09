package utils

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/validator.v9"
)

// utility methods

func BindAndValidate(ctx *gin.Context, obj interface{}) error {

	var validate *validator.Validate

	// Bind URI Params
	if err := ctx.ShouldBindUri(obj); err != nil {
		return err
	}

	// Bind Query Params
	if err := ctx.ShouldBindQuery(obj); err != nil {
		return err
	}

	// Bind JSON Body
	if hasJSONTag(obj) {
		if err := ctx.ShouldBindJSON(obj); err != nil {
			return err
		}
	}

	validate = validator.New()
	// Validate the struct
	if err := validate.Struct(obj); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			var errorMessages []string
			for _, err := range errs {
				fieldName := err.Field()
				tag := err.Tag()
				errorMessages = append(errorMessages,
					strings.Join([]string{fieldName, "is required"}, " ")+" ("+tag+")")
			}
			return gin.Error{
				Err:  errs,
				Type: http.StatusBadRequest,
				Meta: errorMessages,
			}
		}
		return err
	}

	return nil
}

func hasJSONTag(obj interface{}) bool {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return false
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if _, ok := field.Tag.Lookup("json"); ok {
			return true
		}
	}

	return false
}
