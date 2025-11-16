# Container + Ingress Template Scaffold

Spin up a Deployment, Service, and Ingress for any container image. Useful for exposing a simple API or web application publicly.

## Custom Resource

- Kind: `ContainerIngress`
- Group/Version: `templates.stolos.cloud/v1`

Spec fields:

| Field | Type | Description |
| --- | --- | --- |
| `image` | string | Container image to deploy (required). |
| `replicas` | int32 | Pod replica count (default `1`). |
| `containerPort` | int32 | Port exposed by the container (default `8080`). |
| `host` | string | Fully-qualified domain to publish via Ingress (required). |
| `path` | string | HTTP path prefix for the Ingress (default `/`). |
| `tlsSecretName` | string | Optional TLS secret name for HTTPS. |

## Local smoke test

Create `test.yaml`:

```yaml
apiVersion: templates.stolos.cloud/v1
kind: ContainerIngress
metadata:
  name: api
  namespace: default
spec:
  image: ghcr.io/example/api:latest
  host: api.example.com
```

Run the flight locally:

```bash
go run ./cmd/main < test.yaml
```

The flight prints a JSON array containing the Deployment, Service, and Ingress.
