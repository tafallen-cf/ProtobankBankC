# Kubernetes Deployment

Kubernetes manifests for deploying the Auth Service to a Kubernetes cluster.

## Prerequisites

- Kubernetes cluster (1.24+)
- kubectl configured
- kustomize (optional, but recommended)
- cert-manager (for TLS certificates)
- nginx-ingress-controller

## Resources

- **deployment.yaml**: Main application deployment with 3 replicas
- **service.yaml**: ClusterIP service exposing port 80
- **configmap.yaml**: Non-sensitive configuration
- **secret.yaml**: Sensitive credentials (JWT secret, DB password)
- **hpa.yaml**: Horizontal Pod Autoscaler (3-10 replicas)
- **serviceaccount.yaml**: Service account for the pods
- **ingress.yaml**: Ingress resource for external access
- **kustomization.yaml**: Kustomize configuration

## Quick Start

### 1. Create Namespace

```bash
kubectl create namespace protobank
```

### 2. Update Secrets

**IMPORTANT**: Before deploying, update the secrets in `secret.yaml` with your actual values:

```bash
# Generate base64 encoded secrets
echo -n 'your-db-password' | base64
echo -n 'your-jwt-secret-min-32-chars' | base64

# Edit secret.yaml and replace the placeholder values
```

### 3. Deploy with Kustomize

```bash
kubectl apply -k .
```

Or without kustomize:

```bash
kubectl apply -f serviceaccount.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f hpa.yaml
kubectl apply -f ingress.yaml
```

### 4. Verify Deployment

```bash
# Check pods
kubectl get pods -n protobank -l app=auth-service

# Check service
kubectl get svc -n protobank auth-service

# Check ingress
kubectl get ingress -n protobank auth-service

# Check logs
kubectl logs -n protobank -l app=auth-service --tail=50
```

## Configuration

### Environment Variables

Set via ConfigMap (`configmap.yaml`):
- `DB_HOST`: PostgreSQL host
- `DB_PORT`: PostgreSQL port
- `DB_NAME`: Database name

Set via Secret (`secret.yaml`):
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `JWT_SECRET`: JWT signing secret (min 32 characters)

### Resource Limits

- **Requests**: 100m CPU, 128Mi memory
- **Limits**: 500m CPU, 512Mi memory

Adjust based on your workload in `deployment.yaml`.

### Autoscaling

HPA configuration:
- **Min replicas**: 3
- **Max replicas**: 10
- **Target CPU**: 70%
- **Target Memory**: 80%

## Health Checks

- **Liveness**: `/live` (checks if pod is alive)
- **Readiness**: `/ready` (checks if pod can accept traffic)
- **Startup**: `/health` (checks service health on startup)

## Security

- Runs as non-root user (UID 1000)
- Read-only root filesystem
- Drops all capabilities
- Pod anti-affinity for high availability

## Ingress

Access the service at: `https://api.protobankbankc.com/api/v1/auth`

Update the host in `ingress.yaml` for your domain.

## Monitoring

Prometheus metrics available at `/metrics` endpoint.

Annotations for Prometheus scraping:
```yaml
prometheus.io/scrape: "true"
prometheus.io/port: "8080"
prometheus.io/path: "/metrics"
```

## Troubleshooting

### Pods not starting

```bash
# Check pod events
kubectl describe pod -n protobank -l app=auth-service

# Check logs
kubectl logs -n protobank -l app=auth-service --tail=100
```

### Database connection issues

```bash
# Verify configmap
kubectl get configmap -n protobank auth-service-config -o yaml

# Verify secrets exist
kubectl get secret -n protobank auth-service-secrets

# Test database connectivity
kubectl run -it --rm debug --image=postgres:14 --restart=Never -- psql -h postgres.protobank.svc.cluster.local -U postgres
```

### Ingress not working

```bash
# Check ingress status
kubectl describe ingress -n protobank auth-service

# Verify cert-manager certificate
kubectl get certificate -n protobank auth-service-tls
```

## Scaling

### Manual scaling

```bash
kubectl scale deployment -n protobank auth-service --replicas=5
```

### Disable HPA

```bash
kubectl delete hpa -n protobank auth-service
```

## Cleanup

```bash
kubectl delete -k .
```

Or:

```bash
kubectl delete namespace protobank
```
