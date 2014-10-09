// The jsonhttp package provides general functions for returning
// JSON responses to HTTP requests. It is agnostic about
// the specific form of any returned errors.
package jsonhttp

import (
	"encoding/json"
	"net/http"

	"github.com/juju/errgo"
)

// ErrorToResponse represents a function that can convert a Go error
// into a form that can be returned as a JSON body from an HTTP request.
// The httpStatus value reports the desired HTTP status.
type ErrorToResponse func(err error) (httpStatus int, errorBody interface{})

// ErrorHandler is like http.Handler except it returns an error
// which may be returned as the error body of the response.
// An ErrorHandler function should not itself write to the ResponseWriter
// if it returns an error.
type ErrorHandler func(http.ResponseWriter, *http.Request) error

// HandleErrors returns a function that can be used to convert an ErrorHandler
// into an http.Handler. The given errToResp parameter is used to convert
// any non-nil error returned by handle to the response in the HTTP body.
func HandleErrors(errToResp ErrorToResponse) func(handle ErrorHandler) http.Handler {
	writeError := WriteError(errToResp)
	return func(handle ErrorHandler) http.Handler {
		f := func(w http.ResponseWriter, req *http.Request) {
			if err := handle(w, req); err != nil {
				writeError(w, err)
			}
		}
		return http.HandlerFunc(f)
	}
}

// WriteError returns a function that can be used to write an error to a ResponseWriter
// and set the HTTP status code. The errToResp parameter is used to determine
// the actual error value and status to write.
func WriteError(errToResp ErrorToResponse) func(w http.ResponseWriter, err error) {
	return func(w http.ResponseWriter, err error) {
		status, resp := errToResp(err)
		WriteJSON(w, status, resp)
	}
}

// WriteJSON writes the given value to the ResponseWriter
// and sets the HTTP status to the given code.
func WriteJSON(w http.ResponseWriter, code int, val interface{}) error {
	// TODO consider marshalling directly to w using json.NewEncoder.
	// pro: this will not require a full buffer allocation.
	// con: if there's an error after the first write, it will be lost.
	data, err := json.Marshal(val)
	if err != nil {
		// TODO(rog) log an error if this fails and lose the
		// error return, because most callers will need
		// to do that anyway.
		return errgo.Mask(err)
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
	return nil
}

// JSONHandler is like http.Handler except that it returns a
// body (to be converted to JSON) and an error.
// An ErrorHandler function should not itself write to the
// ResponseWriter.
// TODO(rog) remove ResponseWriter argument from function argument.
// It is redundant (and possibly dangerous) if used in combination with the interface{}
// return.
type JSONHandler func(http.ResponseWriter, *http.Request) (interface{}, error)

// HandleJSON returns a function that can be used to convert an JSONHandler
// into an http.Handler. The given errToResp parameter is used to convert
// any non-nil error returned by handle to the response in the HTTP body
// If it returns a nil value, the original error is returned as a JSON string.
func HandleJSON(errToResp ErrorToResponse) func(handle JSONHandler) http.Handler {
	handleErrors := HandleErrors(errToResp)
	return func(handle JSONHandler) http.Handler {
		f := func(w http.ResponseWriter, req *http.Request) error {
			val, err := handle(w, req)
			if err != nil {
				return errgo.Mask(err, errgo.Any)
			}
			return WriteJSON(w, http.StatusOK, val)
		}
		return handleErrors(f)
	}
}
