#!/bin/sh
# Injects the Umami tracking tag into index.html at container start.
#
# The static SPA is built once into a single image. The Umami script URL and
# website id differ per environment and are not known at build time, so we
# inject them at runtime — the same pattern nginx's own entrypoint uses for
# BACKEND_URL via envsubst.
#
# index.html ships with a marker comment (<!-- UMAMI_ANALYTICS ... -->). When
# both env vars are set we replace that marker with the tracking <script>.
# When they are unset (local dev, preview builds) the marker stays as an inert
# comment and no analytics are loaded.
#
# nginx:1.27-alpine runs every executable /docker-entrypoint.d/*.sh before
# starting nginx, so this runs once per container start.
set -eu

HTML=/usr/share/nginx/html/index.html
MARKER='<!-- UMAMI_ANALYTICS'

if [ -z "${UMAMI_SCRIPT_URL:-}" ] || [ -z "${UMAMI_WEBSITE_ID:-}" ]; then
  echo "umami: UMAMI_SCRIPT_URL/UMAMI_WEBSITE_ID not set — analytics disabled"
  exit 0
fi

if [ ! -f "$HTML" ]; then
  echo "umami: $HTML not found — skipping injection" >&2
  exit 0
fi

TAG="<script defer src=\"${UMAMI_SCRIPT_URL}\" data-website-id=\"${UMAMI_WEBSITE_ID}\"></script>"

# '|' delimiter is safe: URLs contain '/' but not '|'. Replace the whole
# marker comment line with the tracking tag.
sed -i "s|${MARKER}.*-->|${TAG}|" "$HTML"

echo "umami: injected tracking tag for website ${UMAMI_WEBSITE_ID}"
