# 13. API-Provisioned Read-Only Datasources

**Problem:** The grafusion provisioner creates datasources via the Grafana REST API (POST /api/datasources). There's no way to make these datasources non-editable by end users -- the `readOnly` field is deliberately excluded from API deserialization (`json:"-"`). Only file-based provisioning supports `editable: false`. Customer org admins can modify or delete datasources that should be managed exclusively by the provisioner.

**Solution:** Two changes:

1. **Expose the `ReadOnly` field** on `AddDataSourceCommand` and `UpdateDataSourceCommand` by changing the JSON tag from `json:"-"` to `json:"readOnly"`. This lets the provisioner send `"readOnly": true` in the API payload.

2. **Allow Grafana Server Admins to bypass the read-only guard.** The three enforcement points (delete-by-ID, delete-by-UID, update) check `ds.ReadOnly && !c.SignedInUser.IsGrafanaAdmin` instead of just `ds.ReadOnly`. The provisioner's `api` account is a Grafana Server Admin; customer accounts are only org admins.

## Result

- The provisioner (Grafana Server Admin) can create, update, and delete read-only datasources
- Customer org admins/editors/viewers are blocked from modifying or deleting read-only datasources
- No changes required to the provisioner -- it already sends `"readOnly": true` in its payloads

## Where to implement

- **`pkg/services/datasources/models.go`:** Change `ReadOnly` JSON tag from `"-"` to `"readOnly"` in both `AddDataSourceCommand` and `UpdateDataSourceCommand`
- **`pkg/api/datasources.go`:** Update the three `if ds.ReadOnly` guards to also check `!c.SignedInUser.IsGrafanaAdmin`