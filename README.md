# Ingress Hub

![logo](/img/logo.png)

Shared Kubernetes Ingress

## Concept

![schema](/img/schema.png)

### Terminal

_Terminal_ has an IP (with service `LoadBalancer` for example) and select on which _gate_ to send traffic regarding `HOST` Header or using SNI.

### Gate

An HTTP endpoint that forward traffic in a _boarding bridge_.

### Boarding Bridge

An HTTP2 secured bidirectionnal tunnel.

### Dock

Engine in cluster which request a _boarding bridge_ to a _terminal_ and propagate ingress `Host` configuration to a _gate_.

## Why

On a cloud privider services type `Loadbalancer` can be expensive when only used for testing cluster.

With _Ingress Hub_ you can share 1 IP for many clusters.

## Usage

### End User

#### Install _ingress-hub_ client

```yaml
```

#### Dock instance

```yaml
apiVersion: ingress-hub.barpilot.io/v1alpha1
kind: Dock
metadata:
  name: example
spec:
  serviceName: ingress
  terminal: "https://myterminal.ingress.hub"
  secret: myterminal
  ingressClass: nginx
---
apiVersion: v1
kind: Secret
metadata:
  name: myterminal
type: kubernetes.io/tls
data:
  tls.crt:XXXX
  tls.key:XXXX
```

#### Ingress

You can use any ingress controller. You just have to set properly the service to dock with.

Traffic is send to an Ingress Controller "as is", so ingress controller is reponsible for SSL (for example).

### Admin Usage

#### Install _ingress-hub_ server

```yaml
```

#### Terminal instance

```yaml
apiVersion: ingress-hub.barpilot.io/v1alpha1
kind: Terminal
metadata:
  name: myterminal
spec:
  serviceName: mylb
```

This create a new service `type: LoadBalancer` and instanciate a deployment for a `terminal`.

#### Gate instance

```yaml
apiVersion: ingress-hub.barpilot.io/v1alpha1
kind: Gate
metadata:
  name: example
spec:
  terminal: myterminal
```

_ingress-hub_ will populate the status with certificates that can be used by a _dock_.

## State

Not done!