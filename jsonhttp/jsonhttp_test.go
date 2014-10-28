// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package jsonhttp_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	gc "gopkg.in/check.v1"
	"gopkg.in/errgo.v1"

	"github.com/juju/utils/jsonhttp"
)

type suite struct{}

var _ = gc.Suite(&suite{})

func (*suite) TestWriteJSON(c *gc.C) {
	rec := httptest.NewRecorder()
	type Number struct {
		N int
	}
	err := jsonhttp.WriteJSON(rec, http.StatusTeapot, Number{1234})
	c.Assert(err, gc.IsNil)
	c.Assert(rec.Code, gc.Equals, http.StatusTeapot)
	c.Assert(rec.Body.String(), gc.Equals, `{"N":1234}`)
	c.Assert(rec.Header().Get("content-type"), gc.Equals, "application/json")
}

var (
	errUnauth = errors.New("unauth")
	errBadReq = errors.New("bad request")
	errOther  = errors.New("other")
	errNil    = errors.New("nil result")
)

type errorResponse struct {
	Message string
}

func errorToResponse(err error) (int, interface{}) {
	resp := &errorResponse{
		Message: err.Error(),
	}
	status := http.StatusInternalServerError
	switch errgo.Cause(err) {
	case errUnauth:
		status = http.StatusUnauthorized
	case errBadReq:
		status = http.StatusBadRequest
	case errNil:
		return status, nil
	}
	return status, &resp
}

var writeErrorTests = []struct {
	err          error
	expectStatus int
	expectResp   *errorResponse
}{{
	err:          errUnauth,
	expectStatus: http.StatusUnauthorized,
	expectResp: &errorResponse{
		Message: errUnauth.Error(),
	},
}, {
	err:          errBadReq,
	expectStatus: http.StatusBadRequest,
	expectResp: &errorResponse{
		Message: errBadReq.Error(),
	},
}, {
	err:          errOther,
	expectStatus: http.StatusInternalServerError,
	expectResp: &errorResponse{
		Message: errOther.Error(),
	},
}, {
	err:          errNil,
	expectStatus: http.StatusInternalServerError,
}}

func (s *suite) TestWriteError(c *gc.C) {
	writeError := jsonhttp.WriteError(errorToResponse)
	for i, test := range writeErrorTests {
		c.Logf("%d: %s", i, test.err)
		rec := httptest.NewRecorder()
		writeError(rec, test.err)
		resp := parseErrorResponse(c, rec.Body.Bytes())
		c.Assert(resp, gc.DeepEquals, test.expectResp)
		c.Assert(rec.Code, gc.Equals, test.expectStatus)
	}
}

func parseErrorResponse(c *gc.C, body []byte) *errorResponse {
	var errResp *errorResponse
	err := json.Unmarshal(body, &errResp)
	c.Assert(err, gc.IsNil)
	return errResp
}

func (s *suite) TestHandleErrors(c *gc.C) {
	handleErrors := jsonhttp.HandleErrors(errorToResponse)

	// Test when handler returns an error.
	handler := handleErrors(func(http.ResponseWriter, *http.Request) error {
		return errUnauth
	})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, new(http.Request))
	c.Assert(rec.Code, gc.Equals, http.StatusUnauthorized)
	resp := parseErrorResponse(c, rec.Body.Bytes())
	c.Assert(resp, gc.DeepEquals, &errorResponse{
		Message: errUnauth.Error(),
	})

	// Test when handler returns nil.
	handler = handleErrors(func(w http.ResponseWriter, _ *http.Request) error {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("something"))
		return nil
	})
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, new(http.Request))
	c.Assert(rec.Code, gc.Equals, http.StatusCreated)
	c.Assert(rec.Body.String(), gc.Equals, "something")
}

var handleErrorsWithErrorAfterWriteHeaderTests = []struct {
	about            string
	causeWriteHeader func(w http.ResponseWriter)
}{{
	about: "write",
	causeWriteHeader: func(w http.ResponseWriter) {
		w.Write([]byte(""))
	},
}, {
	about: "write header",
	causeWriteHeader: func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusOK)
	},
}, {
	about: "flush",
	causeWriteHeader: func(w http.ResponseWriter) {
		w.(http.Flusher).Flush()
	},
}}

func (s *suite) TestHandleErrorsWithErrorAfterWriteHeader(c *gc.C) {
	handleErrors := jsonhttp.HandleErrors(errorToResponse)
	for i, test := range handleErrorsWithErrorAfterWriteHeaderTests {
		c.Logf("test %d: %s", i, test.about)
		handler := handleErrors(func(w http.ResponseWriter, _ *http.Request) error {
			test.causeWriteHeader(w)
			return errgo.New("unexpected")
		})
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, new(http.Request))
		c.Assert(rec.Code, gc.Equals, http.StatusOK)
		c.Assert(rec.Body.String(), gc.Equals, "")
	}
}

func (s *suite) TestHandleJSON(c *gc.C) {
	handleJSON := jsonhttp.HandleJSON(errorToResponse)

	// Test when handler returns an error.
	handler := handleJSON(func(http.Header, *http.Request) (interface{}, error) {
		return nil, errUnauth
	})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, new(http.Request))
	resp := parseErrorResponse(c, rec.Body.Bytes())
	c.Assert(resp, gc.DeepEquals, &errorResponse{
		Message: errUnauth.Error(),
	})
	c.Assert(rec.Code, gc.Equals, http.StatusUnauthorized)

	// Test when handler returns a body.
	handler = handleJSON(func(h http.Header, _ *http.Request) (interface{}, error) {
		h.Set("Some-Header", "value")
		return "something", nil
	})
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, new(http.Request))
	c.Assert(rec.Code, gc.Equals, http.StatusOK)
	c.Assert(rec.Body.String(), gc.Equals, `"something"`)
	c.Assert(rec.Header().Get("Some-Header"), gc.Equals, "value")
}
