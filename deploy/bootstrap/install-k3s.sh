#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "==> Installing k3s (Traefik disabled, klipper servicelb enabled)"
curl -sfL https://get.k3s.io | sh -s - \
  --disable traefik \
  --write-kubeconfig-mode 644

export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

echo "==> Waiting for node to be Ready"
kubectl wait --for=condition=Ready node --all --timeout=120s

echo "==> Installing Helm"
curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

echo "==> Installing ingress-nginx"
helm upgrade --install ingress-nginx ingress-nginx \
  --repo https://kubernetes.github.io/ingress-nginx \
  --namespace ingress-nginx --create-namespace \
  -f "${SCRIPT_DIR}/ingress-nginx-values.yaml" \
  --wait

echo "==> Installing cert-manager"
helm upgrade --install cert-manager cert-manager \
  --repo https://charts.jetstack.io \
  --namespace cert-manager --create-namespace \
  --set crds.enabled=true \
  --wait

echo "==> Applying ClusterIssuers"
kubectl apply -f "${SCRIPT_DIR}/cert-manager-clusterissuer.yaml"

echo "==> Installing ArgoCD"
helm upgrade --install argocd argo-cd \
  --repo https://argoproj.github.io/argo-helm \
  --namespace argocd --create-namespace \
  -f "${SCRIPT_DIR}/argocd-values.yaml" \
  --wait

echo ""
echo "==> Bootstrap complete."
echo ""
echo "ArgoCD initial admin password:"
kubectl -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath="{.data.password}" | base64 -d && echo
echo ""
echo "Access ArgoCD via: kubectl port-forward svc/argocd-server -n argocd 8080:443"
echo ""
echo "Next steps:"
echo "  1. Copy kubeconfig: scp root@89.167.20.94:/etc/rancher/k3s/k3s.yaml ~/.kube/riksdagskollen.yaml"
echo "  2. Apply GHCR pull secret: GHCR_PAT=<token> bash deploy/bootstrap/secrets/ghcr-pull-secret.sh"
echo "  3. Install ARC: see deploy/bootstrap/arc/ README in plan"
echo "  4. Apply ArgoCD Application: kubectl apply -f deploy/argocd/application.yaml"
