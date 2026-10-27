# XDatabase Composition

**Composition Name:** database-pipeline  
**Composite Type:** example.org/v1alpha1/XDatabase  
**Mode:** Pipeline

## Pipeline Steps

| Step | Function |
|------|----------|
| patch-and-transform | function-patch-and-transform |
| ready | function-auto-ready |

Pipeline functions are not executed. Only static input.resources entries are listed below; functions may create additional resources at runtime.


## Managed Resources

Static resource definitions found: 1.

| Resource Name | Kind | API Version |
|---------------|------|-------------|
| database | Instance | rds.aws.upbound.io/v1beta1 |


## Field Mappings

### database (Instance)

FromComposite patches copy from the composite to the resource; ToComposite patches copy from the resource to the composite.

| Patch Type | Source Field | Destination Field | Transformation |
|------------|--------------|-------------------|----------------|
| FromCompositeFieldPath | spec.parameters.region | spec.forProvider.region | Direct copy |
| FromCompositeFieldPath | spec.parameters.size | spec.forProvider.allocatedStorage | [{"math":{"multiply":2,"type":"Multiply"},"type":"math"}] |
| ToCompositeFieldPath | status.atProvider.endpoint | status.connection.endpoint | Direct copy |
| CombineFromComposite | metadata.name, spec.parameters.region | metadata.annotations[example.org/name] | %s-%s |




