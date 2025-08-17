# HashPost Kubernetes Deployment

This directory contains Kubernetes deployment configurations for HashPost using Kustomize for environment-specific customizations.

## Structure

```
deploy/
├── base/                    # Base Kubernetes manifests
│   ├── kustomization.yaml   # Base Kustomize configuration
│   ├── namespace.yaml       # Namespace definition
│   ├── configmap.yaml       # Application configuration
│   ├── secret.yaml          # Secrets (default values)
│   ├── postgres-*.yaml      # PostgreSQL database
│   ├── backend-*.yaml       # HashPost backend API
│   ├── frontend-*.yaml      # HashPost frontend UI
│   └── ingress.yaml         # Ingress configuration
└── overlays/
    ├── testing/             # Testing environment
    └── production/          # Production environment
```

## Prerequisites

Before deploying HashPost, ensure you have:

1. **Kubernetes cluster** (DigitalOcean Kubernetes recommended)
2. **kubectl** configured to access your cluster
3. **Kustomize** installed (`kubectl kustomize` or standalone)
4. **Container images** pushed to DigitalOcean Container Registry

## Required Secrets

### GitHub Repository Secrets

Configure these secrets in your GitHub repository settings:

- `REGISTRY_NAME`: Your DigitalOcean Container Registry name
- `CLUSTER_NAME`: Your DigitalOcean Kubernetes cluster name  
- `DIGITALOCEAN_ACCESS_TOKEN`: DigitalOcean API token with registry and Kubernetes access

### Kubernetes Secrets

Before deploying, create the IBE keys secret:

```bash
# Generate IBE keys (run locally)
make setup-ibe-keys

# Create Kubernetes secret with IBE keys
kubectl create secret generic hashpost-ibe-keys \
  --from-file=keys/ \
  --namespace=hashpost-testing  # or hashpost-production
```

### Production Secrets

For production, update the secrets with secure values:

```bash
# Create production secrets
kubectl create secret generic hashpost-secrets \
  --from-literal=DB_USER=your_db_user \
  --from-literal=DB_PASSWORD=your_secure_db_password \
  --from-literal=JWT_SECRET=your_super_secure_jwt_secret \
  --from-literal=POSTGRES_USER=your_db_user \
  --from-literal=POSTGRES_PASSWORD=your_secure_db_password \
  --from-literal=POSTGRES_DB=hashpost \
  --namespace=hashpost-production
```

## Deployment

### Testing Environment

Deploys automatically on pushes to `main` branch:

```bash
# Manual deployment
kubectl apply -k deploy/overlays/testing
```

### Production Environment

Deploys automatically on GitHub releases:

```bash
# Manual deployment
kubectl apply -k deploy/overlays/production
```

## Configuration

### Environment Variables

Key configuration is managed through ConfigMaps and Secrets:

- **ConfigMap** (`hashpost-config`): Non-sensitive configuration
- **Secret** (`hashpost-secrets`): Sensitive data (passwords, tokens)

### Customization

Each overlay can customize:

- **Replica counts**: Scale deployments up/down
- **Resource limits**: CPU and memory allocation
- **Domain names**: Ingress hostnames
- **Environment-specific settings**: CORS origins, debug flags

### Image Updates

The GitHub Actions workflow automatically:

1. Builds container images
2. Tags with appropriate versions (release tag or commit SHA)
3. Pushes to DigitalOcean Container Registry
4. Updates Kustomization files with new image tags
5. Deploys to Kubernetes using images from DigitalOcean Container Registry

## Monitoring

Check deployment status:

```bash
# Get all resources
kubectl get all -n hashpost-testing  # or hashpost-production

# Check pod logs
kubectl logs -f deployment/hashpost-backend-testing -n hashpost-testing

# Check ingress
kubectl get ingress -n hashpost-testing
```

## Scaling

Scale deployments as needed:

```bash
# Scale backend
kubectl scale deployment hashpost-backend-production --replicas=5 -n hashpost-production

# Scale frontend  
kubectl scale deployment hashpost-frontend-production --replicas=3 -n hashpost-production
```

## Troubleshooting

### Common Issues

1. **ImagePullBackOff**: Check registry credentials and image names
2. **Pending Pods**: Check resource requests vs available cluster resources
3. **Database Connection**: Verify database is running and secrets are correct
4. **Ingress Issues**: Check ingress controller and DNS configuration

### Debug Commands

```bash
# Describe failing pods
kubectl describe pod <pod-name> -n <namespace>

# Check events
kubectl get events -n <namespace> --sort-by='.lastTimestamp'

# Port forward for local testing
kubectl port-forward service/hashpost-backend 8888:8888 -n <namespace>
```

## Security Notes

- All containers run as non-root users
- Secrets are base64 encoded (change default values!)
- Network policies should be configured for production
- Consider using external secret management (e.g., External Secrets Operator)
- Regular security scanning of container images recommended

## Registry Information

Images are pushed to DigitalOcean Container Registry with the following naming convention:

- **Testing**: `<commit-sha>` (e.g., `abc1234`) - triggered by pushes to main
- **Production**: `<release-tag>` (e.g., `v1.0.0`) - triggered by GitHub releases

The DigitalOcean Kubernetes cluster has integrated registry access, so no additional image pull secrets are needed.
