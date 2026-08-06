#!/usr/bin/env sh
set -eu

# Rebuild the released revision and restart the two application services.
docker compose up -d --build --remove-orphans
docker image prune -f
