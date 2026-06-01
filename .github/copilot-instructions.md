# Recycler — Copilot Instructions

## Build & Lint (Required Before Finishing)

Always run both of the following before marking a task complete. Fix any errors before responding.

```bash
go build ./...
./bin/golangci-lint run ./...
```

## Project Layout

- `internal/controller/` — reconciler logic (`recycler_controller.go`, `monitor_controller.go`) and custom metrics (`metrics.go`)
- `api/v1alpha1/` — CRD types; run `make generate manifests` after changing types
- `test/e2e/` — end-to-end tests (Ginkgo); helpers live in `test/utils/utils.go`
- `config/` — Kustomize overlays; `config/overlays/debug/` is used by e2e tests via `make deploy-debug`
- `helm-charts/recycler/` — Helm chart (generated via helmify)

## Key Conventions

- Controllers use `retry.RetryOnConflict` for all status/annotation updates
- Pod metrics history storage is configurable: `memory` or `annotation` (see `StorageMemory` / `StorageAnnotation` constants)
- Custom Prometheus metrics are registered in `internal/controller/metrics.go` via `metrics.Registry` — no changes to `cmd/main.go` needed
- e2e tests talk to the cluster via `kubectl` exec; metrics assertions use `utils.FetchControllerMetrics` + `utils.MetricValue` (plain-text Prometheus format, no JSON)
