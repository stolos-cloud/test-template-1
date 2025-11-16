# Base (Empty) Template Scaffold

This scaffold provides a bare-minimum starter workspace for creating a Template.

## Usage

1. Fill in AirwayInputs.yml
2. Customize code
3. Customize documentation
4. Check the status of your deployed Template CRD in the "Templates" section of Stolos UI.

## Local smoke test

Create a k8s CustomResource `test.yaml` which respects the Spec you defined in the code and run:

```
go run ./cmd/main < test.yaml
```

## CICD Pipeline

This template is compiled automatically when changes are detected.