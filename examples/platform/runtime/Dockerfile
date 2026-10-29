FROM xpkg.upbound.io/crossplane-contrib/function-go-templating:v0.13.2@sha256:20198cb6ea5b4b8a48042343410ec9031e5ec43431d874364ecd3eac8e861290 AS go-files
COPY networks.gotmpl /templates/networks.gotmpl

FROM xpkg.upbound.io/crossplane-contrib/function-kcl:v0.12.3@sha256:601b48f70f03a87ef67c75c14434fb3851950a9f33ae621e376f4179978400b2 AS kcl-files
COPY main.k /templates/main.k
