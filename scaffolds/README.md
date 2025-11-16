# Template Scaffolds

This folder contains the scaffolds that are shown in the UI when creating a template.

## Pre-defined scaffolds:

Generally, the pre-defined scaffolds should not be modified. Instead, you can define your own scaffold following the same structure.

### Base (Empty)

This is the basic scaffold which contains no resources except the Flight and Airway definitions. Generally this should not be modified.

### Basic container deployment

Provides a Deployment-only starter that exposes a container via the Service creation logic in Go.

### Container + Ingress

Adds a Kubernetes Ingress layer on top of the basic container Deployment.

### Container + Ingress + DB

Extends the previous scaffold with a CloudNativePG PostgreSQL cluster and helpful env vars injected into the Deployment.

### Container + Ingress + DB + Redis

Adds a cache tier (Redis or Valkey container + Service) alongside the app + Postgres resources.

### Full Stack

Builds the API (container + ingress + DB + Redis) plus a static nginx frontend with its own ConfigMap, Service, and Ingress.

### ...

...
