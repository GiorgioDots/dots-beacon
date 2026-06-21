package auth

// User is the authenticated identity extracted from a verified Keycloak token.
type User struct {
	Subject       string   // "sub" — stable unique user id
	Username      string   // "preferred_username"
	Email         string   // "email"
	EmailVerified bool     // "email_verified"
	Name          string   // "name"
	Roles         []string // realm roles ("realm_access.roles")
}

// claims mirrors the relevant fields of a Keycloak access/ID token payload.
type claims struct {
	Subject           string `json:"sub"`
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	Name              string `json:"name"`
	RealmAccess       struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
}

func (c claims) toUser() User {
	return User{
		Subject:       c.Subject,
		Username:      c.PreferredUsername,
		Email:         c.Email,
		EmailVerified: c.EmailVerified,
		Name:          c.Name,
		Roles:         c.RealmAccess.Roles,
	}
}

// HasRole reports whether the user has the given Keycloak realm role.
func (u User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}
