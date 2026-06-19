# Runbook: self-hosted Umami analytics

Umami gives privacy-friendly, **cookieless** visitor counts (no consent banner
needed) without sending data to a third party. It runs as a self-hosted pod
with its own dedicated Postgres, behind `analytics.<host>`. The frontend loads
a tiny tracking script only when configured.

Everything is gated behind `umami.enabled` in the Helm chart — **off by
default**. Enabling it is a deliberate, two-deploy process (the website UUID
only exists after the first deploy).

## What gets deployed (when `umami.enabled=true`)

| Object | Purpose |
|---|---|
| `…-umami-db` StatefulSet + headless Service | Dedicated Postgres (PVC) |
| `…-umami` Deployment + Service | Umami app, port 3000 |
| `…-umami` Ingress | `analytics.<host>` over TLS |
| Frontend env `UMAMI_SCRIPT_URL`/`UMAMI_WEBSITE_ID` | Injected tracking tag |

## How the tracking tag works

The SPA is built once into a single image, so the script URL and website id are
**not** baked in. `frontend/index.html` ships an inert marker comment. At
container start `frontend/docker-entrypoint.d/40-umami-analytics.sh` replaces
that marker with `<script defer src=… data-website-id=…>` — but only when both
env vars are set. Unset (local dev, preview) = no tag, no analytics.

## Prerequisites — pre-create two secrets

The chart references secrets by name; it never creates them (same pattern as
`dbSecretName`). Create both in the release namespace before enabling:

```sh
# 1. Umami's Postgres password
kubectl create secret generic riksdagskollen-umami-db-secret \
  --from-literal=POSTGRES_PASSWORD="$(openssl rand -hex 24)"

# 2. Umami app: DATABASE_URL must match the password above and point at the
#    umami-db service; APP_SECRET signs Umami sessions.
DB_PW="<same password as above>"
kubectl create secret generic riksdagskollen-umami-secret \
  --from-literal=DATABASE_URL="postgresql://umami:${DB_PW}@riksdagskollen-umami-db:5432/umami" \
  --from-literal=APP_SECRET="$(openssl rand -hex 32)"
```

## Enable — first deploy (brings up Umami)

Set in your prod values overlay:

```yaml
umami:
  enabled: true
  ingress:
    host: analytics.riksdagskollen.example.com   # real host
```

Deploy. Umami runs its own DB migrations on startup. Then:

1. Open `https://analytics.<host>`, log in with the default `admin` / `umami`
   account and **immediately change the password**.
2. Add the website (Settings → Websites → Add). Domain = the public site host
   (`riksdagskollen.example.com`).
3. Copy the **Website ID** (UUID) and the tracking **script URL**
   (`https://analytics.<host>/script.js`).

## Wire the tag — second deploy

```yaml
umami:
  frontend:
    scriptUrl: "https://analytics.riksdagskollen.example.com/script.js"
    websiteId: "<uuid-from-step-3>"
```

Redeploy. The frontend pod restarts and the entrypoint injects the tag.

## Verify

```sh
# Tag present in served HTML
curl -s https://riksdagskollen.example.com/ | grep -o 'data-website-id="[^"]*"'

# Umami app healthy
kubectl exec deploy/riksdagskollen-umami -- wget -qO- localhost:3000/api/heartbeat
```

Then load the public site in a browser and confirm a hit appears on the Umami
dashboard (realtime view).

## Notes

- Cookieless by design — no GDPR consent banner required for basic analytics.
- Disabling (`umami.enabled=false`) removes all Umami objects but **not** the
  PVC or the secrets; delete those manually if decommissioning for good.
- The frontend image is analytics-agnostic: with the env vars unset it serves
  exactly as before.
