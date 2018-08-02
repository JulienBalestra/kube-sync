# kube-sync [![CircleCI](https://circleci.com/gh/JulienBalestra/kube-sync.svg?style=svg)](https://circleci.com/gh/JulienBalestra/kube-sync) [![Docker Repository on Quay](https://quay.io/repository/julienbalestra/kube-sync/status "Docker Repository on Quay")](https://quay.io/repository/julienbalestra/kube-sync)

kube-sync synchronise (create/update) a configmap from a source namespace to all namespaces.

Have a look to the [docs](docs) and the [examples](examples).

kube-sync also exposes the following [metrics](docs/metrics.csv).

kube-sync annotate the configmap with the references of the source and the update timestamp.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    kube-sync/source: '{"namespace":"kube-system","name":"to-sync","uid":"2ba4f600-883f-11e8-ae10-42010a10e004","resourceVersion":"47067967","last-update":1533203794}'
```
