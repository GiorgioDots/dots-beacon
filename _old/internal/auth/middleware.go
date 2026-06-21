package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// contextKey is the gin context key under which the authenticated User is stored.
const contextKey = "auth.user"

// Middleware returns a gin handler that requires a valid Keycloak bearer token.
// On success the User is stored in the context (read it with UserFromContext);
// on failure it responds 401 and aborts the chain.
func (a *Authenticator) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := bearerToken(c.Request.Header.Get("Authorization"))
		if !ok {
			unauthorized(c, "missing bearer token")
			return
		}

		user, err := a.Verify(c.Request.Context(), raw)
		if err != nil {
			unauthorized(c, "invalid or expired token")
			return
		}

		c.Set(contextKey, user)
		c.Next()
	}
}

// UserFromContext returns the authenticated user set by Middleware.
func UserFromContext(c *gin.Context) (User, bool) {
	v, ok := c.Get(contextKey)
	if !ok {
		return User{}, false
	}
	user, ok := v.(User)
	return user, ok
}

// bearerToken extracts the token from an "Authorization: Bearer <token>" header.
func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	return token, token != ""
}

func unauthorized(c *gin.Context, msg string) {
	c.Header("WWW-Authenticate", `Bearer error="invalid_token"`)
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": msg})
}
