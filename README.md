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

```text
./kube-sync kube-system to-sync --kubeconfig-path ~/.kube/config -v 1
I0804 14:20:26.812652   12678 kubesync.go:139] Starting to sync source cm/to-sync from ns kube-system ...
I0804 14:20:26.827481   12678 kubesync.go:166] Annotate the destination configmaps with the reference of the source kube-sync/source: {"namespace":"kube-system","name":"to-sync","uid":"c63cc178-97d4-11e8-9192-5404a66983a9","resourceVersion":"1039","last-update":1533385226}
I0804 14:20:26.827650   12678 kubesync.go:172] The configmap to sync across 4 namespaces is: {"metadata":{"name":"to-sync","namespace":"kube-system","selfLink":"/api/v1/namespaces/kube-system/configmaps/to-sync","uid":"c63cc178-97d4-11e8-9192-5404a66983a9","resourceVersion":"1039","creationTimestamp":"2018-08-04T10:54:34Z"},"data":{"bar":"two","foo":"one"}}
I0804 14:20:26.829784   12678 kubesync.go:193] Successfully sync cm/to-sync from ns kube-system to the ns default
I0804 14:20:26.831985   12678 kubesync.go:193] Successfully sync cm/to-sync from ns kube-system to the ns kube-public
I0804 14:20:26.832001   12678 kubesync.go:178] Skipping sync over the namespace kube-system: namespace of the source configmap
I0804 14:20:26.833722   12678 kubesync.go:193] Successfully sync cm/to-sync from ns kube-system to the ns ns-8d7f7740-1911-440b-b31c-102eb904d167
I0804 14:20:26.835222   12678 kubesync.go:193] Successfully sync cm/to-sync from ns kube-system to the ns ns-91e8b569-376a-4fe2-9a45-da7bde6e066a
I0804 14:20:26.835238   12678 kubesync.go:210] Successfully sync in 22.596424ms
I0804 14:20:26.835269   12678 kubesync.go:229] Starting prometheus exporter on 0.0.0.0:8484/metrics
I0804 14:20:26.835331   12678 kubesync.go:246] Starting pprof on 127.0.0.1:6060/debug/pprof
I0804 14:20:26.835341   12678 kubesync.go:267] Starting to sync every 1m0s
I0804 14:21:26.835443   12678 kubesync.go:139] Starting to sync source cm/to-sync from ns kube-system ...
I0804 14:21:26.840726   12678 kubesync.go:166] Annotate the destination configmaps with the reference of the source kube-sync/source: {"namespace":"kube-system","name":"to-sync","uid":"c63cc178-97d4-11e8-9192-5404a66983a9","resourceVersion":"1039","last-update":1533385286}
I0804 14:21:26.840787   12678 kubesync.go:172] The configmap to sync across 4 namespaces is: {"metadata":{"name":"to-sync","namespace":"kube-system","selfLink":"/api/v1/namespaces/kube-system/configmaps/to-sync","uid":"c63cc178-97d4-11e8-9192-5404a66983a9","resourceVersion":"1039","creationTimestamp":"2018-08-04T10:54:34Z"},"data":{"bar":"two","foo":"one"}}
I0804 14:21:26.844151   12678 kubesync.go:193] Successfully sync cm/to-sync from ns kube-system to the ns default
I0804 14:21:26.847008   12678 kubesync.go:193] Successfully sync cm/to-sync from ns kube-system to the ns kube-public
I0804 14:21:26.847023   12678 kubesync.go:178] Skipping sync over the namespace kube-system: namespace of the source configmap
I0804 14:21:26.849536   12678 kubesync.go:193] Successfully sync cm/to-sync from ns kube-system to the ns ns-8d7f7740-1911-440b-b31c-102eb904d167
I0804 14:21:26.854013   12678 kubesync.go:193] Successfully sync cm/to-sync from ns kube-system to the ns ns-91e8b569-376a-4fe2-9a45-da7bde6e066a
I0804 14:21:26.854032   12678 kubesync.go:210] Successfully sync in 18.603709ms
```
