# XDatabase

A database API for application teams.

**API Group:** example.org  
**API Version:** v1alpha1  
**Kind:** XDatabase  
**Claim Kind:** Database  

## Spec Fields

| Name | Type | Description | Required | Default | Constraints |
|------|------|-------------|----------|---------|-------------|
| parameters | object |  | ✅ | - | - |


## Status Fields

| Name | Type | Description |
|------|------|-------------|
| connection | object |  |


## Example

```yaml
apiVersion: example.org/v1alpha1
kind: Database
metadata:
  name: example
spec:
  # Add your spec fields here
```
