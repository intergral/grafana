# 2. Auth Proxy Org Context (X-WEBAUTH-ORG)

**Problem:** When Grafana sits behind a reverse proxy that handles authentication, the proxy can pass user identity (email, username) via headers. But there's no way to tell Grafana which organization the user should be placed in. In a multi-tenant setup, users end up in the wrong org.

**Solution:** Add a new auth proxy header field `OrgName` mapped to a configurable header (default `X-WEBAUTH-ORG`). When present, Grafana resolves the org name to an org ID and uses it as the user's active organization.

## How it works

- The auth proxy client configuration gains an `OrgName` header mapping (alongside existing `Name`, `Email`, `Login`, `Groups`, `Role` fields).
- When processing an auth proxy request, the proxy client and the Grafana authn client both need access to an `orgService` to resolve org names to IDs.
- On authentication, if the `OrgName` header is present, look up the org by name and set `identity.OrgID`.
- In the user sync hook, when the request's org ID is 0 (unset), prefer the org ID from the identity (which came from the auth proxy header) rather than defaulting to org 1.

## Configuration

Changes in `[auth.proxy]` section of `defaults.ini`:

| Key | Change | Value |
|-----|--------|-------|
| `header_name` | default changed | `X-WEBAUTH-EMAIL` |
| `header_property` | default changed | `email` |
| `auto_sign_up` | default changed | `false` |
| `sync_ttl` | default changed | `0` |
| `headers` | added mapping | `OrgName:X-WEBAUTH-ORG` |

## Where to implement

- **Auth proxy client:** add org name field, inject orgService, resolve name to ID
- **Grafana authn client:** same orgService injection and resolution
- **Authn registration/wiring:** pass orgService to both clients
- **User sync:** respect identity.OrgID over request.OrgID when request has no org set