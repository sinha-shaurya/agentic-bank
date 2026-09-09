package responsehandler

import "github.com/sinha-shaurya/agentic-bank/utils/errors"


const (
	StatusFailure = 0
	StatusSuccess = 1
)

const defaultSuccessMessage = "success"

type Response struct {
	Status    int              `json:"status"`
	Message   string           `json:"message"`
	Data      any              `json:"data,omitempty"`
	ErrorCode errors.ErrorCode `json:"error_code,omitempty"`
}
