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
    kube-sync: '{"ns":"kube-system","cm":"to-sync","uuid":"e00f2250-9624-11e8-95b1-5404a66983a9","ts":1533196912}'
```
