# exips
A sidekick for the ingress controller that exposes the external IPs of Kubernetes nodes as a Service, designed for setups where the ingress controller binds directly to host ports 80 and 443 and no additional IP addresses or load balancer are available.

Only IPs of schedulable nodes in ready state are exposed; control-plane IPs are excluded if taints prevent workload scheduling.

# Deploy

The `deploy` directory contains Kubernetes Objects and a [Kustomize](https://kustomize.io/) configuration.
It can be deployed using the [Flux CD `GitRepository`API](https://fluxcd.io/flux/components/kustomize/kustomizations/).

Containers for linux/amd64 and linux/arm64 are available on Docker Hub: [fabiant7t/exips](https://hub.docker.com/r/fabiant7t/exips).

## Traefik Ingress
If you use Traefik for HTTP(S) routing and run `exips` with the default configuration, enable `publishedService` and set it to `exips/exips`:

```yaml
values:
  providers:
    kubernetesIngress:
      publishedService:
        enabled: true
        pathOverride: "exips/exips"
```

### Traefik example
The following example shows a Traefik deployment using the `exips` Service in the `exips` namespace (default) as the source for external IP addresses:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  labels:
    pod-security.kubernetes.io/enforce: privileged
  name: traefik

---
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: traefik
  namespace: traefik
spec:
  interval: 24h
  url: https://traefik.github.io/charts

---
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: traefik
  namespace: traefik
spec:
  chart:
    spec:
      chart: traefik
      interval: 12h
      sourceRef:
        kind: HelmRepository
        name: traefik
        namespace: traefik
      version: 39.0.0
  interval: 30m
  values:
    deployment:
      enabled: true
      kind: DaemonSet
    ingressClass:
      enabled: true
      isDefaultClass: false
    metrics:
      prometheus:
        serviceMonitor:
          enabled: true
    ports:
      web:
        expose:
          default: true
        exposedPort: 80
        hostPort: 80
      websecure:
        expose:
          default: true
        exposedPort: 443
        hostPort: 443
    providers:
      kubernetesGateway:
        enabled: true
      kubernetesIngress:
        allowExternalNameServices: true
        publishedService:
          enabled: true
          pathOverride: exips/exips
    service:
      enabled: false
    updateStrategy:
      rollingUpdate:
        maxSurge: 0
        maxUnavailable: 1
      type: RollingUpdate
```
