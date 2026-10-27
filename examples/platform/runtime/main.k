oxr = option("params").oxr
items = [{
    apiVersion = "v1"
    kind = "ConfigMap"
    metadata = {
        name = "${oxr.metadata.name}-${n.name}"
        annotations = {"krm.kcl.dev/composition-resource-name" = "network-${i}"}
    }
    data = {region = oxr.spec.region, cidr = n.cidr}
} for i, n in oxr.spec.networks]
