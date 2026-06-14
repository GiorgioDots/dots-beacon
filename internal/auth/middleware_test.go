package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBearerToken(t *testing.T) {
	cases := []struct {
		header  string
		want    string
		wantOK  bool
	}{
		{"Bearer abc.def.ghi", "abc.def.ghi", true},
		{"bearer abc", "abc", true}, // case-insensitive scheme
		{"Bearer   ", "", false},    // empty token
		{"Token abc", "", false},    // wrong scheme
		{"", "", false},             // no header
		{"abc.def.ghi", "", false},  // no scheme
	}
	for _, tc := range cases {
		got, ok := bearerToken(tc.header)
		if got != tc.want || ok != tc.wantOK {
			t.Errorf("bearerToken(%q) = (%q, %v), want (%q, %v)", tc.header, got, ok, tc.want, tc.wantOK)
		}
	}
}

// Middleware must reject requests with no/invalid Authorization header before it
// ever touches the verifier, so these run without a live Keycloak.
func TestMiddlewareRejectsWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	a := &Authenticator{} // verifier unused on these paths

	for _, header := range []string{"", "Token xyz", "Bearer "} {
		r := gin.New()
		r.Use(a.Middleware())
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Authorization=%q: status = %d, want 401", header, w.Code)
		}
	}
}
