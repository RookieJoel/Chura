#!/bin/sh
set -eu

# Pattern A: Kong resolves upstreams (*.service.consul) via Consul DNS on :8600.
# dns_resolver must be a literal IP:port (not "consul:8600").
CONSUL_IP="$(getent ahosts consul | awk '{print $1; exit}')"
if [ -z "${CONSUL_IP}" ]; then
  echo "kong entrypoint: failed to resolve consul on the compose network" >&2
  exit 1
fi
export KONG_DNS_RESOLVER="${CONSUL_IP}:8600"
echo "kong entrypoint: using KONG_DNS_RESOLVER=${KONG_DNS_RESOLVER}"

exec /docker-entrypoint.sh kong docker-start
