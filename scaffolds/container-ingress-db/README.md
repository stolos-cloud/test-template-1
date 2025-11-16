# Container + Ingress + Database Template Scaffold

Extends the container + ingress setup with a managed PostgreSQL database powered by CloudNativePG (CNPG).

## Custom Resource

- Kind: `ContainerIngressDB`
- Group/Version: `templates.stolos.cloud/v1`

Spec overview:

| Field | Type | Description |
| --- | --- | --- |
| `image` | string | Backend container image (required). |
| `replicas` | int32 | Backend replicas (default `2`). |
| `containerPort` | int32 | Container port (default `8080`). |
| `host` / `path` / `tlsSecretName` | string | Ingress settings (`host` required). |
| `database.clusterName` | string | Name for the CNPG cluster (required). |
| `database.databaseName` | string | Database to bootstrap (required). |
| `database.instances` | int32 | CNPG instances (default `1`). |
| `database.storageSize` | string | Persistent volume size (default `10Gi`). |
| `database.postgresVersion` | string | Major version (default `16`). |

The generated Deployment includes env vars (`DATABASE_HOST`, `DATABASE_NAME`, `DATABASE_PORT`) that point at the CNPG cluster RW service.

## Local smoke test

```yaml
apiVersion: templates.stolos.cloud/v1
kind: ContainerIngressDB
metadata:
  name: api-with-db
  namespace: default
spec:
  image: ghcr.io/example/api:latest
  host: api.example.com
  database:
    clusterName: api-db
    databaseName: app
```

```bash
go run ./cmd/main < test.yaml
```
