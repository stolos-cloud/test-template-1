# Full Stack Template Scaffold

Provision a complete stack that includes:

- Backend API Deployment with Service + Ingress
- PostgreSQL (CloudNativePG) cluster
- Cache tier (Redis or Valkey)
- Static frontend (nginx) with ConfigMap-provided HTML and its own Service + Ingress

## Custom Resource

- Kind: `FullStack`
- Group/Version: `templates.stolos.cloud/v1`

### Backend spec (`spec.backend`)

| Field | Description |
| --- | --- |
| `image` | Backend container image (required). |
| `replicas` | Default `2`. |
| `containerPort` | Default `8080`. |
| `host` / `path` / `tlsSecretName` | Ingress properties (host required). |

### Database spec (`spec.database`)

Same as the Container + Ingress + DB scaffold (CNPG cluster settings).

### Cache spec (`spec.cache`)

| Field | Description |
| --- | --- |
| `flavor` | `redis` (default) or `valkey`. |
| `port` | Default `6379`. |

### Frontend spec (`spec.frontend`)

| Field | Description |
| --- | --- |
| `host` / `path` / `tlsSecretName` | Ingress config for the static site (host required). |
| `image` | nginx image (default `nginx:stable-alpine`). |
| `replicas` | Default `1`. |
| `staticContent` | Optional inline HTML for `index.html`. When omitted, a helper page pointing to the backend host is generated.

## Local smoke test

```yaml
apiVersion: templates.stolos.cloud/v1
kind: FullStack
metadata:
  name: storefront
  namespace: default
spec:
  backend:
    image: ghcr.io/example/api:latest
    host: api.example.com
  frontend:
    host: app.example.com
  database:
    clusterName: storefront-db
    databaseName: app
```

```bash
go run ./cmd/main < test.yaml
```
