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
| &nbsp;&nbsp;↳ region | string | Deployment region. | ✅ | - | Allowed: `eu-west-1`, `eu-central-1` |
| &nbsp;&nbsp;↳ size | integer | Storage in GiB. | ✅ | `20` | Allowed: `20`, `50`, `100`, Min: 20, Max: 100 |
| &nbsp;&nbsp;↳ backup | boolean |  | ❌ | `false` | - |
| &nbsp;&nbsp;↳ labels | object |  | ❌ | `{"team":"platform"}` | - |
| &nbsp;&nbsp;↳ note | string | Choose primary \| replica.<br>Use `region` to place the database. | ❌ | - | - |
| &nbsp;&nbsp;↳ replicas | integer |  | ❌ | `0` | - |
| &nbsp;&nbsp;↳ zones | list(string) |  | ❌ | `["eu-west-1a"]` | MinItems: 1, MaxItems: 3 |


## Status Fields

| Name | Type | Description |
|------|------|-------------|
| connection | object |  |
| &nbsp;&nbsp;↳ endpoint | string |  |
| &nbsp;&nbsp;↳ port | integer |  |


## Example

```yaml
apiVersion: example.org/v1alpha1
kind: Database
metadata:
  name: example
spec:
  # Add your spec fields here
```
