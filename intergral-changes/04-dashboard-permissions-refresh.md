# 4. Dashboard Permissions Refresh After Save

**Problem:** After saving a brand-new dashboard, Grafana redirects to the new URL. But the permission caches (both backend access control and frontend user permissions) haven't been updated yet, so the redirect fails with a 403 "not allowed to view" error. This is worse in multi-instance deployments where the redirect may hit a different instance that hasn't propagated the new dashboard yet.

**Solution:** Two-pronged fix:

## Frontend

- After saving, call `contextSrv.fetchUserPermissions()` to refresh the frontend's permission cache before redirecting.
- Add an `afterSave=1` query parameter to the redirect URL.
- When loading a dashboard with `afterSave=1`, if the load returns 404/403/500, retry up to 6 times with a delay. Strip the param after successful load.

## Backend

- When clearing user permission cache, also clear basic role caches so managed role permissions for newly created resources are picked up immediately.
- Add a method to clear cached scope resolution for specific resources.

## Where to implement

- **Frontend save hooks:** refresh permissions + add afterSave param
- **Dashboard page state manager:** retry logic on load when afterSave is present
- **Backend access control service:** extend cache clearing to cover basic roles and scope resolution
- **API client:** allow passing options (like `showErrorAlert: false`) to subresource calls so retries don't spam error toasts