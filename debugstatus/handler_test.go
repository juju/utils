// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package debugstatus_test

import (
	"encoding/json"
	"net/http"

	jc "github.com/juju/testing/checkers"
	"github.com/juju/testing/httptesting"
	"github.com/juju/utils/debugstatus"
	"github.com/julienschmidt/httprouter"
	gc "gopkg.in/check.v1"
	"gopkg.in/errgo.v1"

	"github.com/juju/httprequest"
)

var errorMapper httprequest.ErrorMapper = func(err error) (httpStatus int, errorBody interface{}) {
	return http.StatusInternalServerError, httprequest.RemoteError{
		Message: err.Error(),
	}
}

type handlerSuite struct {
}

var _ = gc.Suite(&handlerSuite{})

var errUnauthorized = errgo.New("you shall not pass!")

func newHTTPHandler(h *debugstatus.Handler) http.Handler {
	errMapper := httprequest.ErrorMapper(func(err error) (httpStatus int, errorBody interface{}) {
		code, status := "", http.StatusInternalServerError
		switch err {
		case errUnauthorized:
			code, status = "unauthorized", http.StatusUnauthorized
		case debugstatus.ErrNoPprofConfigured:
			code, status = "forbidden", http.StatusForbidden
		}
		return status, httprequest.RemoteError{
			Code:    code,
			Message: err.Error(),
		}
	})

	handlers := errMapper.Handlers(func(httprequest.Params) (*debugstatus.Handler, error) {
		return h, nil
	})
	r := httprouter.New()
	for _, h := range handlers {
		r.Handle(h.Method, h.Path, h.Handle)
	}
	return r
}

func (s *handlerSuite) TestServeDebugStatus(c *gc.C) {
	httpHandler := newHTTPHandler(&debugstatus.Handler{
		Check: func() map[string]debugstatus.CheckResult {
			return debugstatus.Check(debugstatus.ServerStartTime)
		},
	})
	httptesting.AssertJSONCall(c, httptesting.JSONCallParams{
		Handler: httpHandler,
		URL:     "/debug/status",
		ExpectBody: httptesting.BodyAsserter(func(c *gc.C, body json.RawMessage) {
			var result map[string]debugstatus.CheckResult
			err := json.Unmarshal(body, &result)
			c.Assert(err, gc.IsNil)
			for k, v := range result {
				v.Duration = 0
				result[k] = v
			}
			c.Assert(result, jc.DeepEquals, map[string]debugstatus.CheckResult{
				"server_started": {
					Name:   "Server started",
					Value:  debugstatus.StartTime.String(),
					Passed: true,
				},
			})
		}),
	})
}

func (s *handlerSuite) TestServeDebugStatusWithNilCheck(c *gc.C) {
	httpHandler := newHTTPHandler(&debugstatus.Handler{})
	httptesting.AssertJSONCall(c, httptesting.JSONCallParams{
		Handler:    httpHandler,
		URL:        "/debug/status",
		ExpectBody: map[string]debugstatus.CheckResult{},
	})
}

func (s *handlerSuite) TestServeDebugInfo(c *gc.C) {
	version := debugstatus.Version{
		GitCommit: "some-git-status",
		Version:   "a-version",
	}
	httpHandler := newHTTPHandler(&debugstatus.Handler{
		Version: version,
	})
	httptesting.AssertJSONCall(c, httptesting.JSONCallParams{
		Handler:      httpHandler,
		URL:          "/debug/info",
		ExpectStatus: http.StatusOK,
		ExpectBody:   version,
	})
}

var debugPprofPaths = []string{
	"/debug/pprof/",
	"/debug/pprof/cmdline",
	"/debug/pprof/profile?seconds=1",
	"/debug/pprof/symbol",
	"/debug/pprof/goroutine",
}

func (s *handlerSuite) TestServeDebugPprof(c *gc.C) {
	httpHandler := newHTTPHandler(&debugstatus.Handler{
		CheckPprofAllowed: func(req *http.Request) error {
			if req.Header.Get("Authorization") == "" {
				return errUnauthorized
			}
			return nil
		},
	})
	authHeader := make(http.Header)
	authHeader.Set("Authorization", "let me in")
	for i, path := range debugPprofPaths {
		c.Logf("%d. %s", i, path)
		httptesting.AssertJSONCall(c, httptesting.JSONCallParams{
			Handler:      httpHandler,
			URL:          path,
			ExpectStatus: http.StatusUnauthorized,
			ExpectBody: httprequest.RemoteError{
				Code:    "unauthorized",
				Message: "you shall not pass!",
			},
		})
		rr := httptesting.DoRequest(c, httptesting.DoRequestParams{
			Handler: httpHandler,
			URL:     path,
			Header:  authHeader,
		})
		c.Assert(rr.Code, gc.Equals, http.StatusOK)
	}
}

func (s *handlerSuite) TestDebugPprofForbiddenWhenNotConfigured(c *gc.C) {
	httpHandler := newHTTPHandler(&debugstatus.Handler{})
	httptesting.AssertJSONCall(c, httptesting.JSONCallParams{
		Handler:      httpHandler,
		URL:          "/debug/pprof/",
		ExpectStatus: http.StatusForbidden,
		ExpectBody: httprequest.RemoteError{
			Code:    "forbidden",
			Message: "no pprof access configured",
		},
	})
}
