# k8s-test
A simple application for testing a k8s environment.

## Prepare

1. Install ingress

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

helm install -f helm/ingress/values.yaml ingress-nginx ingress-nginx/ingress-nginx
```

2. Add repository `bitnamy` for postgres

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
```