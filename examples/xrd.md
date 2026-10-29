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
| spec.parameters.region | string | Deployment region. | true | - | {&#34;enum&#34;:[&#34;eu-west-1&#34;,&#34;eu-central-1&#34;]} |
| spec.parameters.size | integer | Storage in GiB. | true | 20 | {&#34;enum&#34;:[20,50,100],&#34;maximum&#34;:100,&#34;minimum&#34;:20} |
| spec.parameters.backup | boolean |  | false | false | {} |
| spec.parameters.labels | object |  | false | {&#34;team&#34;:&#34;platform&#34;} | {} |
| spec.parameters.note | string | Choose primary \| replica.<br>Use `region` to place the database.<br> | false | - | {} |
| spec.parameters.replicas | integer |  | false | 0 | {} |
| spec.parameters.zones | list(string) |  | false | [&#34;eu-west-1a&#34;] | {&#34;maxItems&#34;:3,&#34;minItems&#34;:1} |
| spec.parameters.zones[] | string |  | false | - | {} |

<a id="status-fields-5"></a>
#### Status Fields

| path | type | description | required | default | constraints |
| --- | --- | --- | --- | --- | --- |
| status.connection | object |  | false | - | {} |
| status.connection.endpoint | string |  | false | - | {} |
| status.connection.port | integer |  | false | - | {} |
