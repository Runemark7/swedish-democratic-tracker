#!/usr/bin/env bash
# Creates a GHCR image pull secret in the riksdagskollen namespace.
# Run locally after copying kubeconfig from the VPS.
#
# Requires:
#   - KUBECONFIG pointing at the k3s cluster
#   - GHCR_PAT env var: GitHub classic PAT with read:packages scope
#
# Usage:
#   GHCR_PAT=ghp_xxxx bash deploy/bootstrap/secrets/ghcr-pull-secret.sh
set -euo pipefail

: "${GHCR_PAT:?GHCR_PAT env var is required}"

kubectl create namespace riksdagskollen --dry-run=client -o yaml | kubectl apply -f -

kubectl create secret docker-registry ghcr-pull-secret \
  --namespace riksdagskollen \
  --docker-server=ghcr.io \
  --docker-username=runemark7 \
  --docker-password="${GHCR_PAT}" \
  --dry-run=client -o yaml | kubectl apply -f -

echo "ghcr-pull-secret created in namespace riksdagskollen"
