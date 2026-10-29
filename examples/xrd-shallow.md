<a id="crossplane-api-reference-1"></a>
# Crossplane API reference

<a id="api-definitions-2"></a>
## API definitions

<a id="xdatabase-v1alpha1-3"></a>
### XDatabase · v1alpha1

A database API for application teams.

| Property | Value |
| --- | --- |
| API Group | example.org |
| API Version | v1alpha1 |
| Kind | XDatabase |
| Scope |  |
| Claim Kind | Database |
| Served | true |
| Referenceable | true |

<a id="spec-fields-4"></a>
#### Spec Fields

| path | type | description | required | default | constraints |
| --- | --- | --- | --- | --- | --- |
| spec.parameters | object |  | true | - | {} |

<a id="status-fields-5"></a>
#### Status Fields

| path | type | description | required | default | constraints |
| --- | --- | --- | --- | --- | --- |
| status.connection | object |  | false | - | {} |
