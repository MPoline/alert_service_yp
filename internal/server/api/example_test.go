package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MPoline/alert_service_yp/internal/server/api"
)

func ExampleInitRouter() {
	router := api.InitRouter()

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/ping")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	println("Response status:", resp.Status)
}

func TestRouterEndpoints(t *testing.T) {
	router := api.InitRouter()

	testCases := []struct {
		method string
		path   string
		status int
	}{
		{"GET", "/ping", http.StatusOK},
		{"GET", "/", http.StatusOK},
		{"GET", "/value/gauge/test_metric", http.StatusNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tc.method, tc.path, nil)
			router.ServeHTTP(w, req)

			if w.Code != tc.status {
				t.Errorf("Expected status %d, got %d", tc.status, w.Code)
			}
		})
	}
}
