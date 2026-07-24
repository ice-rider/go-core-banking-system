package errors

import (
	"net/http"

	pkgerrors "go-core-banking-system/pkg/errors"
)

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

func MapError(err error) (int, ErrorResponse) {
	if err == nil {
		return http.StatusOK, ErrorResponse{}
	}

	status, code := pkgerrors.ToHTTPStatus(err)
	return status, ErrorResponse{
		Error: err.Error(),
		Code:  code,
	}
}
