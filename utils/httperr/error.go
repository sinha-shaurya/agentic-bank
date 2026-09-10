// Package httperr defines the service→controller error channel: an HTTP
// status and a wire error code carried alongside the cause.
package httperr

import "net/http"

// ErrorCode is the machine-readable discriminator on the wire.
type ErrorCode string

const (
	// Request shape and auth.
	ErrValidation   ErrorCode = "err_validation"
	ErrUnauthorized ErrorCode = "err_unauthorized"
	ErrForbidden    ErrorCode = "err_forbidden"
	ErrNotFound     ErrorCode = "err_not_found"

	// Ledger outcomes.
	ErrIdempotencyKeyConflict ErrorCode = "err_idempotency_key_conflict"
	ErrInsufficientFunds      ErrorCode = "err_insufficient_funds"

	ErrInternalServer ErrorCode = "err_internal_server"
)

// Definition binds a wire error code to the HTTP status and the default
// client-visible message served for it.
type Definition struct {
	Code       ErrorCode
	StatusCode int
	Message    string
}

// catalog is the single source of truth for code→status. A new failure mode is
// a new line here, not a status picked at the call site.
var catalog = []Definition{
	{ErrValidation, http.StatusBadRequest, "request is invalid"},
	{ErrUnauthorized, http.StatusUnauthorized, "authentication required"},
	{ErrForbidden, http.StatusForbidden, "not permitted"},
	{ErrNotFound, http.StatusNotFound, "resource not found"},
	{ErrIdempotencyKeyConflict, http.StatusConflict, "idempotency key was already used with a different request"},
	{ErrInsufficientFunds, http.StatusUnprocessableEntity, "insufficient funds"},
	{ErrInternalServer, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)},
}

var definitions = indexCatalog()

// indexCatalog panics on a duplicate code: two definitions for one wire code
// means one of them is silently unreachable, and startup is when to find out.
func indexCatalog() map[ErrorCode]Definition {
	index := make(map[ErrorCode]Definition, len(catalog))
	for _, definition := range catalog {
		if _, duplicate := index[definition.Code]; duplicate {
			panic("httperr: duplicate error code in catalog: " + string(definition.Code))
		}
		index[definition.Code] = definition
	}
	return index
}

// Lookup returns the catalog entry registered for errorCode.
func Lookup(errorCode ErrorCode) (Definition, bool) {
	definition, found := definitions[errorCode]
	return definition, found
}

// HttpError carries an HTTP status, a wire error code, and the cause. It
// implements error and unwraps to Err.
