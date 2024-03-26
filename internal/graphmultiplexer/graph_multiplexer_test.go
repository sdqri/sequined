package graphmultiplexer_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	gm "github.com/sdqri/sequined/internal/graphmultiplexer"
	hr "github.com/sdqri/sequined/internal/hyperrenderer"
	"github.com/sdqri/sequined/internal/observer"
	"github.com/stretchr/testify/assert"
)

func init() {
	if err := os.Chdir("../.."); err != nil {
		panic(err)
	}
}

type RequestTestCase struct {
	req                *http.Request
	ExpectedStatusCode int
	ExpectedBody       []byte
}

func TestHandleGraphHttpRequest(t *testing.T) {
	testCases := []struct {
		name             string
		rootGenerator    func() *hr.Webpage
		requestTestCases []RequestTestCase
	}{
		{
			name: "single node",
			rootGenerator: func() *hr.Webpage {
				return hr.NewWebpage(hr.WebpageTypeHub)
			},
			requestTestCases: []RequestTestCase{
				{
					req:                httptest.NewRequest(http.MethodGet, "/", strings.NewReader("")),
					ExpectedStatusCode: http.StatusOK,
				},
			},
		},
		{
			name: "single node with prefix",
			rootGenerator: func() *hr.Webpage {
				return hr.NewWebpage(hr.WebpageTypeHub, hr.WithPathPrefix("/testprefix"))
			},
			requestTestCases: []RequestTestCase{
				{
					req:                httptest.NewRequest(http.MethodGet, "/testprefix", strings.NewReader("")),
					ExpectedStatusCode: http.StatusOK,
				},
				{
					req:                httptest.NewRequest(http.MethodGet, "/", strings.NewReader("")),
					ExpectedStatusCode: http.StatusNotFound,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root := tc.rootGenerator()
			mx, err := gm.New(root)
			assert.NoErrorf(t, err, "Error while creating root")
			for _, rtc := range tc.requestTestCases {
				r := httptest.NewRecorder()
				mx.GraphHandlerFunc(r, rtc.req)
				assert.Equal(t, rtc.ExpectedStatusCode, r.Result().StatusCode)
			}
		})
	}
}

func TestObserverOnMux(t *testing.T) {
	testCases := []struct {
		name string

		rootGenerator              func() *hr.Webpage
		reqs                       []*http.Request
		ExpectedNodeLogLength      int
		ExpectedVisitHistoryLength int
	}{
		{
			name: "root & single visit",
			rootGenerator: func() *hr.Webpage {
				return hr.NewWebpage(hr.WebpageTypeHub)
			},
			reqs: []*http.Request{
				httptest.NewRequest(http.MethodGet, "/", strings.NewReader("")),
			},
			ExpectedNodeLogLength:      1,
			ExpectedVisitHistoryLength: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root := tc.rootGenerator()
			o := observer.New()
			mx, err := gm.New(root, gm.WithObserver(o))
			assert.NoErrorf(t, err, "Error while creating root")
			for _, req := range tc.reqs {
				r := httptest.NewRecorder()
				mx.GraphHandlerFunc(r, req)
			}

			assert.Len(t, mx.Observer.NodeLogMap, tc.ExpectedNodeLogLength)
			assert.Len(t, mx.Observer.VisitHistory, tc.ExpectedVisitHistoryLength)
		})
	}

}
