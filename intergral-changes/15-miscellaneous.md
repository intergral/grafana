# 15. Miscellaneous

- **Command palette simplification:** Reduce the default command palette actions to only what's relevant for FusionReactor
- **Empty session token guard:** Protect against panic when session token is empty in auth service
- **Debug infrastructure:** Docker Compose + nginx/traefik configs for testing auth proxy locally (lives in `debug/` directory)