package errors

import (
	"net/http"
	"strings"
)

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

type httpError struct {
	status int
	code   string
}

var errorMap = []struct {
	pattern string
	handler httpError
}{
	{"account not found", httpError{http.StatusNotFound, "NOT_FOUND"}},
	{"transaction not found", httpError{http.StatusNotFound, "NOT_FOUND"}},
	{"amount must be positive", httpError{http.StatusBadRequest, "VALIDATION_ERROR"}},
	{"owner name is required", httpError{http.StatusBadRequest, "VALIDATION_ERROR"}},
	{"idempotency key is required", httpError{http.StatusBadRequest, "VALIDATION_ERROR"}},
	{"cannot transfer to the same account", httpError{http.StatusBadRequest, "VALIDATION_ERROR"}},
	{"insufficient funds", httpError{http.StatusConflict, "CONFLICT"}},
	{"account is blocked", httpError{http.StatusConflict, "CONFLICT"}},
	{"account is closed", httpError{http.StatusConflict, "CONFLICT"}},
	{"account is not active", httpError{http.StatusConflict, "CONFLICT"}},
	{"balance must be zero to close account", httpError{http.StatusConflict, "CONFLICT"}},
	{"idempotency key already used", httpError{http.StatusConflict, "CONFLICT"}},
}

func MapError(err error) (int, ErrorResponse) {
	if err == nil {
		return http.StatusOK, ErrorResponse{}
	}

	msg := err.Error()
	for _, entry := range errorMap {
		if strings.Contains(msg, entry.pattern) {
			return entry.handler.status, ErrorResponse{
				Error: msg,
				Code:  entry.handler.code,
			}
		}
	}

	return http.StatusInternalServerError, ErrorResponse{
		Error: msg,
		Code:  "INTERNAL_ERROR",
	}
}

func IsDomainError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, entry := range errorMap {
		if strings.Contains(msg, entry.pattern) {
			return true
		}
	}
	return false
}
