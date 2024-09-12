package main

import (
	_ "embed"
	"net/http"
	"net/http/httptest"
	"testing"
)

//go:embed testdata/valid.json
var validTokenRes []byte

//go:embed testdata/invalid.json
var invalidTokenRes []byte

func TestIntegration(t *testing.T) {
	var cases = []struct {
		name        string
		response    []byte
		headers     map[string]string
		wantCode    int
		wantHeaders map[string]string
	}{
		{
			"validToken",
			validTokenRes,
			map[string]string{
				"Authorization": "Bearer magicalmisterytour",
				"X-Tenant-Id":   "em9mdGtvLWxhYg",
			},
			http.StatusNoContent,
			map[string]string{
				"X-Auth-User": "kiwi",
			},
		},
		{
			"invalidToken",
			invalidTokenRes,
			map[string]string{
				"Authorization": "Bearer sadtoken",
			},
			http.StatusUnauthorized,
			nil,
		},
		{
			"invalidTenantId",
			validTokenRes,
			map[string]string{
				"Authorization": "hello",
				"X-Tenant-Id":   "whoami",
			},
			http.StatusUnauthorized,
			nil,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := w.Write(tt.response)
				if err != nil {
					t.Errorf("failed to write body: %s", err)
				}
			}))
			defer ts.Close()

			z := Zitadel(zitadelOpts{
				Host:         ts.URL,
				VerifyTenant: true,
			})

			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatalf("failed to create request %s", err)
			}
			for name, val := range tt.headers {
				req.Header.Set(name, val)
			}
			res := httptest.NewRecorder()
			z.ServeHTTP(res, req)

			if res.Code != tt.wantCode {
				t.Errorf("unexpected http code, want %d, got %d", tt.wantCode, res.Code)
			}
			for name, val := range tt.wantHeaders {
				got := res.Header().Get(name)
				if got != val {
					t.Errorf("unexpected http header %s, want %s, got %s", name, val, got)
				}
			}
		})
	}
}
