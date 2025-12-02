# Kustomize Overlays

This directory contains Kustomize overlays for different deployment environments.

## Usage

### Default (Base Configuration)
Deploy with default settings (no debug logging):
```bash
kubectl apply -k config/default
```

### Debug Overlay
Deploy with debug logging enabled:
```bash
kubectl apply -k config/overlays/debug
```

This enables:
- `--zap-log-level=1` (debug level)
- `--zap-devel=true` (development mode with stack traces)

### Production Overlay
Deploy with production-ready settings:
```bash
kubectl apply -k config/overlays/production
```

This configures:
- `--zap-log-level=info` (info level logging)
- 2 replicas for high availability
- Increased resource limits
- No development mode

## Helm Chart Integration

If using the Helm chart, you can reference these overlays or add similar configuration options to `values.yaml`.

## Creating Custom Overlays

To create a custom overlay:
1. Create a new directory under `config/overlays/`
2. Add a `kustomization.yaml` that references `../../default`
3. Add patch files for your specific needs
4. Apply with `kubectl apply -k config/overlays/your-overlay`
