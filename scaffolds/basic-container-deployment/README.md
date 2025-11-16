# Basic Container Deployment Template Scaffold

This scaffold provides a minimal Template that deploys a single container in Kubernetes using an `apps/v1.Deployment`.

## Custom Resource

The generated Custom Resource has:

- Kind: `ContainerDeployment`
- Group / Version (in the CR spec): `templates.stolos.cloud/v1`

Spec fields:

- `image` (string, required): Container image to run.
- `replicas` (int32, optional, default: `1`): Number of pod replicas.
- `port` (int32, optional, default: `80`): Container port to expose.

## Usage

1. Adjust `AirwayInputs.yml` to suit your naming preferences if desired.
2. Customize the spec type or Deployment generation logic in the Go code if you need additional fields.
3. After your Template is built and deployed, create instances of `ContainerDeployment` to roll out simple container workloads.

## Local smoke test

Create a k8s CustomResource `test.yaml` that respects the `ContainerDeployment` spec and run:

```bash
go run ./cmd/main < test.yaml
```

The program will output a JSON array containing a single `apps/v1.Deployment` resource.

