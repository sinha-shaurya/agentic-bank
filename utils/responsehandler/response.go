// Package responsehandler owns the single API response envelope.
package responsehandler

import "github.com/sinha-shaurya/agentic-bank/utils/httperr"

// Envelope-level outcome flags, distinct from the HTTP status.
const (
	StatusFailure = 0
	StatusSuccess = 1
)

const defaultSuccessMessage = "success"

// Response is the envelope every endpoint returns. Status carries no omitempty
// on purpose: StatusFailure is the zero value.
type Response struct {
	Status    int               `json:"status"`
	Message   string            `json:"message"`
	Data      any               `json:"data"`
	ErrorCode httperr.ErrorCode `json:"error_code"`
}
