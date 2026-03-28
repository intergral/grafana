# Debug Auth Proxy Infrastructure

This directory contains a Docker Compose setup for testing the Auth Proxy OrgName feature locally.

## Architecture

- **Traefik**: Reverse proxy that routes `grafana.localhost` to Grafana on port 3000
- **nginx-auth**: Simulates an authentication proxy that returns auth headers
- **Grafana**: Your local Grafana instance (running on host port 3000)

## Setup

1. Start the infrastructure:
   ```bash
   cd debug
   docker-compose up -d
   ```

2. Start Grafana locally on port 3000:
   ```bash
   make run
   ```

3. Access Grafana through the proxy:
   ```
   http://grafana.localhost
   ```

## Configuration

### Auth Headers (nginx/default.conf)

The nginx container returns these headers:
- `X-WEBAUTH-NAME`: "Test User"
- `X-WEBAUTH-EMAIL`: "test.user@example.com"
- `X-WEBAUTH-ORG`: "admin@localhost"
- `X-WEBAUTH-USER`: "testuser"
- `X-WEBAUTH-ROLE`: "Admin" (commented out by default)

### Grafana Configuration

Make sure your `conf/custom.ini` or `conf/defaults.ini` has:

```ini
[auth.proxy]
enabled = true
header_name = X-WEBAUTH-EMAIL
header_property = email
auto_sign_up = false
sync_ttl = 0
whitelist =
headers = "Name:X-WEBAUTH-NAME Role:X-WEBAUTH-ROLE Email:X-WEBAUTH-EMAIL OrgName:X-WEBAUTH-ORG"
```

## Testing OrgName Feature

1. Create an organization in Grafana with name "admin@localhost"
2. Access through `http://grafana.localhost`
3. User should be automatically assigned to the "admin@localhost" organization

## Cleanup

```bash
docker-compose down
```

## Troubleshooting

- **Can't access grafana.localhost**: Add `127.0.0.1 grafana.localhost` to `/etc/hosts`
- **Headers not forwarded**: Check Traefik dashboard at `http://localhost:8080`
- **Grafana not reachable**: Ensure Grafana is running on host port 3000 and `172.17.0.1` is accessible from Docker
