package utils

import "reflect"

func CheckIfStructFieldsAreNil(s interface{}) bool {
    v := reflect.ValueOf(s)
    if v.Kind() != reflect.Ptr || v.IsNil() {
        panic("expected a pointer to a struct")
    }

    v = v.Elem()
    if v.Kind() != reflect.Struct {
        panic("expected a pointer to a struct")
    }

    for i := 0; i < v.NumField(); i++ {
        field := v.Field(i)
        if !isZero(field) {
            return false
        }
    }

    return true
}

// isZero checks if a reflect.Value is considered zero.
func isZero(v reflect.Value) bool {
    switch v.Kind() {
    case reflect.Ptr, reflect.Interface:
        return v.IsNil()
    case reflect.Slice, reflect.Map:
        return v.IsNil() || v.Len() == 0
    case reflect.Array:
        return v.Len() == 0
    case reflect.Struct:
        return v.IsZero()
    default:
        return v.IsZero()
    }
}