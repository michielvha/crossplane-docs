<a id="crossplane-api-reference-1"></a>
# Crossplane API reference

<a id="compositions-2"></a>
## Compositions

<a id="database-pipeline-3"></a>
### database-pipeline

example.org/v1alpha1 / XDatabase · Pipeline

<a id="pipeline-steps-4"></a>
#### Pipeline steps

| Step | Function | Type | Source | Package |
| --- | --- | --- | --- | --- |
| patch-and-transform | function-patch-and-transform | patch-and-transform | Static resource templates and patches |  |
| ready | function-auto-ready | custom |  |  |

Declared definitions are a source reference. Functions may create additional resources at runtime; see the named render scenarios for actual results.

<a id="declared-resources-5"></a>
#### Declared resources

| Resource | Kind | API version | Evidence |
| --- | --- | --- | --- |
| database | Instance | rds.aws.upbound.io/v1beta1 | declared |

<a id="field-mappings-6"></a>
#### Field Mappings

<a id="database-7"></a>
##### database

| Patch type | Source field | Destination field | Transformation |
| --- | --- | --- | --- |
| FromCompositeFieldPath | spec.parameters.region | spec.forProvider.region | Direct copy |
| FromCompositeFieldPath | spec.parameters.size | spec.forProvider.allocatedStorage | Apply math: {&#34;multiply&#34;:2,&#34;type&#34;:&#34;Multiply&#34;} |
| ToCompositeFieldPath | status.atProvider.endpoint | status.connection.endpoint | Direct copy |
| CombineFromComposite | metadata.name, spec.parameters.region | metadata.annotations[example.org/name] | Combine string %s-%s |

<a id="patch-and-transform-source-8"></a>
#### patch-and-transform source

```yaml
apiVersion: pt.fn.crossplane.io/v1beta1
kind: Resources
resources:
    - base:
        apiVersion: rds.aws.upbound.io/v1beta1
        kind: Instance
      name: database
      patches:
        - fromFieldPath: spec.parameters.region
          toFieldPath: spec.forProvider.region
        - fromFieldPath: spec.parameters.size
          toFieldPath: spec.forProvider.allocatedStorage
          transforms:
            - math:
                multiply: 2
                type: Multiply
              type: math
          type: FromCompositeFieldPath
        - fromFieldPath: status.atProvider.endpoint
          toFieldPath: status.connection.endpoint
          type: ToCompositeFieldPath
        - combine:
            strategy: string
            string:
                fmt: '%s-%s'
            variables:
                - fromFieldPath: metadata.name
                - fromFieldPath: spec.parameters.region
          toFieldPath: metadata.annotations[example.org/name]
          type: CombineFromComposite
```
