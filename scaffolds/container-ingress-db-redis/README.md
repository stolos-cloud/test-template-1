# Container + Ingress + Database + Redis Template Scaffold

Adds a managed PostgreSQL cluster (CNPG) and a cache tier (Redis or Valkey) alongside the baseline container + ingress deployment.

## Custom Resource

- Kind: `ContainerIngressDBRedis`
- Group/Version: `templates.stolos.cloud/v1`

Key spec sections:

| Field | Description |
| --- | --- |
| `image` / `replicas` / `containerPort` | Backend deployment configuration. |
| `host` / `path` / `tlsSecretName` | Ingress exposure (host required). |
| `database.*` | Same knobs as the `Container + Ingress + DB` scaffold. |
| `cache.flavor` | `redis` (default) or `valkey`. |
| `cache.port` | Cache service port (default `6379`). |

The backend Deployment exports env vars for both the PostgreSQL RW service and the cache Service.

## Local smoke test

```yaml
apiVersion: templates.stolos.cloud/v1
kind: ContainerIngressDBRedis
metadata:
  name: api-suite
  namespace: default
spec:
  image: ghcr.io/example/api:latest
  host: api.example.com
  database:
    clusterName: api-suite-db
    databaseName: app
  cache:
    flavor: valkey
```

```bash
go run ./cmd/main < test.yaml
```
