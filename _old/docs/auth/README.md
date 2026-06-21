# Authentication (Keycloak)

dots-beacon uses **Keycloak** as the identity provider. Keycloak owns login and
user management; the APIs are **resource servers** that only verify the bearer
access token on each request. Token validation is shared across services via the
[`internal/auth`](../../internal/auth/) package, so adding auth to a service is a
couple of env vars and one middleware.

## How it works

```
 Browser / client                Keycloak (:8081)                 API (:8080)
 ────────────────                ────────────────                 ───────────
 login / get token  ───────────► realm: dots-beacon
                                  issues JWT access token
 request + Authorization: Bearer <token> ───────────────────────► auth.Middleware()
                                  ◄── JWKS (cached) ───────────────  verify sig/iss/exp
                                                                     → User in context
```

The API never sees passwords. It fetches Keycloak's signing keys (JWKS) via OIDC
discovery once at startup, caches them, and verifies each token's **signature,
issuer, and expiry** locally — no per-request call to Keycloak.

## Local dev stack

`docker compose` runs Keycloak in dev mode (H2 storage, persisted to the
`keycloak-data` volume) with the **`dots-beacon` realm pre-seeded** from
[`keycloak/realm.json`](../../keycloak/realm.json) on first start:

| Thing | Value |
|-------|-------|
| Keycloak URL | http://localhost:8081 |
| Master admin console | http://localhost:8081/admin/master/console → `KEYCLOAK_ADMIN_USER` / `KEYCLOAK_ADMIN_PASSWORD` (dev: `admin` / `admin`) — manages **all** realms |
| Realm admin console | http://localhost:8081/admin/dots-beacon/console → realm `admin` / `admin` — manages the `dots-beacon` realm only |
| Realm | `dots-beacon` |
| Issuer | `http://localhost:8081/realms/dots-beacon` |
| Client | `dots-beacon-app` (public, direct access grants enabled) |
| Test user | `dev` / `dev` (email `dev@dots-beacon.local`, realm role `user`) |
| Realm admin user | `admin` / `admin` (realm roles `admin`, `user`; `realm-management` → `realm-admin`, so it can add/manage users) |
| Realm roles | `user`, `admin` |

> Keycloak listens on host **8081** (its container port 8080 is remapped) because
> the API uses 8080.

### Managing users

Two ways to add/manage users:

- **Realm admin** `admin` / `admin` — log in at
  http://localhost:8081/admin/dots-beacon/console. Scoped to the `dots-beacon`
  realm via the built-in `realm-management` → `realm-admin` role. Use this for
  day-to-day user management.
- **Master admin** (`KEYCLOAK_ADMIN_USER`) — log in at
  http://localhost:8081/admin/master/console. Superuser over every realm.

To seed users permanently, add them to
[`keycloak/realm.json`](../../keycloak/realm.json) (re-import requires wiping the
volume — see troubleshooting). Users created through the console live in the
`keycloak-data` volume and persist across restarts.

Bring it up (via `task`, or directly):

```bash
docker compose --env-file .env.local up -d --wait
```

## Getting a token (dev)

The pre-seeded client allows the password grant for quick testing:

```bash
TOKEN=$(curl -s -X POST \
  http://localhost:8081/realms/dots-beacon/protocol/openid-connect/token \
  -d grant_type=password -d client_id=dots-beacon-app \
  -d username=dev -d password=dev | jq -r .access_token)

curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/me
# {"user":{"Subject":"...","Username":"dev","Email":"dev@dots-beacon.local",
#          "EmailVerified":true,"Name":"Dev User","Roles":["user"]}}
```

> The password grant is for **local testing only**. Real clients use the
> Authorization Code flow against Keycloak; the API side is identical — it just
> validates whatever bearer token arrives.

## Using it in a service

See [`internal/auth`](../../internal/auth/). Minimal wiring (as in
[`apps/api/cmd/main.go`](../../apps/api/cmd/main.go)):

```go
authCfg := auth.ConfigFromEnv()          // KEYCLOAK_ISSUER_URL / KEYCLOAK_CLIENT_ID
protected := r.Group("/")
if authCfg.Enabled() {                    // empty issuer => auth disabled
    a, err := auth.New(ctx, authCfg)      // OIDC discovery (needs Keycloak reachable)
    if err != nil { /* fatal */ }
    protected.Use(a.Middleware())
}
protected.GET("/me", func(c *gin.Context) {
    user, _ := auth.UserFromContext(c)    // populated by the middleware
    c.JSON(200, gin.H{"user": user})
})
```

API surface:

| Symbol | Purpose |
|--------|---------|
| `auth.ConfigFromEnv()` / `auth.Config` | Read issuer + client id from env |
| `auth.New(ctx, cfg)` | Discover provider, build verifier |
| `(*Authenticator).Middleware()` | gin middleware: require valid token, store `User` |
| `auth.UserFromContext(c)` | Get the authenticated `User` |
| `User.HasRole("admin")` | Check a Keycloak realm role |

Authorization (enforcing roles/permissions) is intentionally **not** here — this
package only authenticates. Build role/permission checks on top of `User.Roles`.

## Configuration

| Variable | Meaning |
|----------|---------|
| `KEYCLOAK_ISSUER_URL` | Realm issuer URL. **Empty disables auth** (dev convenience). |
| `KEYCLOAK_CLIENT_ID` | Expected token audience (`aud`). Empty ⇒ audience check skipped (still verifies sig/iss/exp). |
| `KEYCLOAK_ADMIN_USER` / `KEYCLOAK_ADMIN_PASSWORD` | Keycloak bootstrap admin (dev only). |

### Audience note

Keycloak access tokens have **no `aud`** for the requesting client by default, so
a plain setup would fail audience validation. This realm therefore includes an
**audience mapper** on `dots-beacon-app` (see
[`keycloak/realm.json`](../../keycloak/realm.json)) that puts `dots-beacon-app`
into the access token's `aud`. With `KEYCLOAK_CLIENT_ID=dots-beacon-app` set, the
API enforces that audience — so a token must have been minted for this app, not
just any client in the realm.

To **disable** audience checking instead, leave `KEYCLOAK_CLIENT_ID` empty; the
signature, issuer, and expiry are still verified. If you change the mapper or
client id, the two must agree or you'll get `expected audience … got []`.

### Issuer note

The token's `iss` must exactly match `KEYCLOAK_ISSUER_URL` (including host and
port). In dev mode Keycloak derives the issuer from the request URL, so a token
fetched at `localhost:8081` has issuer `…localhost:8081…` — keep the API's
`KEYCLOAK_ISSUER_URL` aligned. If the API runs **inside** compose it would reach
Keycloak at `http://keycloak:8080`, which mismatches the `localhost:8081` issuer
in tokens; fix by setting a fixed `KC_HOSTNAME` on Keycloak or using
`Config.SkipIssuerCheck` (last resort).

## Troubleshooting

| Symptom | Cause / fix |
|---------|-------------|
| API fatal: `bind: address already in use` on `:8080` | Keycloak also defaulting to 8080 — it's mapped to host **8081** here; keep them distinct. |
| API fatal at startup: discovery error | Keycloak not up yet / wrong `KEYCLOAK_ISSUER_URL`. `task` waits for Keycloak healthy first. |
| 401 `invalid or expired token` with a fresh token | Issuer mismatch (see above) or token expired (`accessTokenLifespan` 300s in the dev realm). |
| Realm/user missing after edits to `realm.json` | Import runs only on first start. Re-import: `docker volume rm dots-beacon_keycloak-data` then bring Keycloak back up. |
| 404 "Unable to find matching target resource method" | You hit Keycloak, not the API — check the port (API 8080, Keycloak 8081). |
