#!/bin/sh
set -e

# Optional: Kernel-level network packet loss simulation using Linux Traffic Control (tc / netem)
# Set NETEM_LOSS environment variable (e.g. NETEM_LOSS=10% or NETEM_LOSS=30%)
# Note: Requires container capability NET_ADMIN (--cap-add=NET_ADMIN in docker run or compose)
if [ -n "$NETEM_LOSS" ]; then
    echo "[Docker Entrypoint] Applying Linux netem packet loss simulation: ${NETEM_LOSS} on eth0"
    tc qdisc add dev eth0 root netem loss "${NETEM_LOSS}" 2>/dev/null || \
    tc qdisc change dev eth0 root netem loss "${NETEM_LOSS}" 2>/dev/null || \
    echo "[Docker Entrypoint] Warning: Could not configure netem (ensure --cap-add=NET_ADMIN is enabled)"
fi

# Execute the given command or binary
exec "$@"
