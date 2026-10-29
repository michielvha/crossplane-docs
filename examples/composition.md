<a id="crossplane-api-reference-1"></a>
# Crossplane API reference

<a id="compositions-2"></a>
## Compositions

<a id="database-3"></a>
### database

example.org/v1alpha1 / XDatabase · Resources

<a id="declared-resources-4"></a>
#### Declared resources

| Resource | Kind | API version | Evidence |
| --- | --- | --- | --- |
| database | Instance | rds.aws.upbound.io/v1beta1 | legacy |

<a id="field-mappings-5"></a>
#### Field Mappings

<a id="database-6"></a>
##### database

| Patch type | Source field | Destination field | Transformation |
| --- | --- | --- | --- |
| FromCompositeFieldPath | spec.parameters.region | spec.forProvider.region | Direct copy |
| FromCompositeFieldPath | spec.parameters.size | spec.forProvider.allocatedStorage | Apply math: {&#34;multiply&#34;:2,&#34;type&#34;:&#34;Multiply&#34;} |
| ToCompositeFieldPath | status.atProvider.endpoint | status.connection.endpoint | Direct copy |
| CombineFromComposite | metadata.name, spec.parameters.region | metadata.annotations[example.org/name] | Combine string %s-%s |
