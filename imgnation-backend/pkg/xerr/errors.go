package xerr

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/safeblock-dev/werr"
)

var (
	ErrNotImplemented = errors.New("not implemented yet")

	ErrCanceled            = errors.New("canceled")
	ErrDeadlineExceeded    = errors.New("deadline exceeded")
	ErrRemoteServiceFailed = errors.New("remote service failed")

	ErrInvalidAction   = errors.New("invalid action")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrMissingArgument = errors.New("missing argument")
	ErrOutOfRange      = errors.New("out of range")

	ErrPermissionDenied = errors.New("permission denied")
	ErrUnauthorized     = errors.New("unauthorized")

	ErrEntityExists   = errors.New("entity already exists")
	ErrEntityNotFound = errors.New("entity not found")
	ErrEntityOutdated = errors.New("entity outdated")
)

func GetHttpCode(err error) int {
	err = werr.UnwrapAll(err)
	switch {
	case errors.Is(err, ErrNotImplemented):
		return http.StatusNotImplemented
	case errors.Is(err, ErrCanceled), errors.Is(err, ErrDeadlineExceeded):
		return http.StatusRequestTimeout
	case errors.Is(err, ErrRemoteServiceFailed):
		return http.StatusServiceUnavailable
	case
		errors.Is(err, ErrInvalidAction),
		errors.Is(err, ErrInvalidArgument),
		errors.Is(err, ErrOutOfRange),
		errors.Is(err, ErrMissingArgument):
		return http.StatusBadRequest
	case errors.Is(err, ErrPermissionDenied):
		return http.StatusForbidden
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, ErrEntityExists):
		return http.StatusConflict
	case errors.Is(err, ErrEntityNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func HttpHandleError(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	e error,
) {
	err := e
	if werr.IsWrap(err) {
		err = werr.UnwrapAll(e)
	}
	if err == nil {
		return
	}
	log.Println(err.Error())
	switch {
	case errors.Is(err, ErrNotImplemented):
		w.WriteHeader(http.StatusNotImplemented)
	case errors.Is(err, ErrCanceled), errors.Is(err, ErrDeadlineExceeded):
		w.WriteHeader(http.StatusRequestTimeout)
	case errors.Is(err, ErrRemoteServiceFailed):
		log.Println(err.Error())
		w.WriteHeader(http.StatusServiceUnavailable)
	case
		errors.Is(err, ErrInvalidAction),
		errors.Is(err, ErrInvalidArgument),
		errors.Is(err, ErrOutOfRange),
		errors.Is(err, ErrMissingArgument):
		http.Error(w, e.Error(), http.StatusBadRequest)
	case errors.Is(err, ErrPermissionDenied):
		http.Error(w, e.Error(), http.StatusForbidden)
	case errors.Is(err, ErrUnauthorized):
		http.Error(w, e.Error(), http.StatusUnauthorized)
	case errors.Is(err, ErrEntityExists):
		http.Error(w, e.Error(), http.StatusConflict)
	case errors.Is(err, ErrEntityNotFound):
		http.Error(w, e.Error(), http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
