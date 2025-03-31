package httperr

import "fmt"

type StatusError interface {
	StatusCode() int
	Message() string
}

type wrappedWithStatus struct {
	inner      error
	statusCode int
	message    string
}

func ErrWithStatus(status int, msg string, err error) error {
	return &wrappedWithStatus{
		inner:      err,
		statusCode: status,
		message:    msg,
	}
}

func (w *wrappedWithStatus) Error() string {
	return fmt.Sprintf("%s: %s", w.message, w.inner.Error())
}

func (w *wrappedWithStatus) StatusCode() int {
	return w.statusCode
}

func (w *wrappedWithStatus) Message() string {
	return w.message
}
