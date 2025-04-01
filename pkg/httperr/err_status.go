package httperr

import (
	"errors"
	"github.com/mittwald/api-client-go/pkg/httperr"
	"net/http"
)

func StatusForError(err error) int {
	if statusErr := new(StatusError); errors.As(err, statusErr) {
		return (*statusErr).StatusCode()
	}

	if notFoundErr := new(httperr.ErrNotFound); errors.As(err, &notFoundErr) {
		return http.StatusNotFound
	}

	if notFoundErr := new(httperr.ErrPermissionDenied); errors.As(err, &notFoundErr) {
		return http.StatusForbidden
	}

	return http.StatusInternalServerError
}
