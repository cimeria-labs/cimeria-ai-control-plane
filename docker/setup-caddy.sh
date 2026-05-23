#!/bin/bash
# setup-caddy.sh — Provision Caddy reverse proxy on a fresh VM
#
# Prerequisites:
#   - DNS A record for api.cimeria.online pointing to this VM's public IP
#   - Docker and Docker Compose installed
#   - Ports 80 and 443 open in the firewall
#
# Usage:
#   ssh into the VM, then:
#   cd /opt/multica   (or wherever the repo is cloned)
#   bash docker/setup-caddy.sh
#
# This script:
#   1. Installs Caddy Docker plugin (already in compose file)
#   2. Verifies DNS resolution
#   3. Starts the stack with docker compose
#   4. Caddy auto-provisions TLS via Let's Encrypt

set -e

DOMAIN="${BACKEND_DOMAIN:-api.cimeria.online}"
REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"

echo "=== Multica + Caddy Setup ==="
echo "Domain: $DOMAIN"
echo "Repo:   $REPO_DIR"
echo ""

# 1. Check DNS
echo "[1/4] Checking DNS for $DOMAIN..."
PUBLIC_IP=$(curl -sf https://ifconfig.me || curl -sf https://api.ipify.org)
if [ -z "$PUBLIC_IP" ]; then
    echo "WARNING: Could not determine this machine's public IP."
    echo "Make sure $DOMAIN points to this VM's public IP."
else
    RESOLVED_IP=$(dig +short "$DOMAIN" 2>/dev/null | tail -1)
    if [ -z "$RESOLVED_IP" ]; then
        echo "ERROR: $DOMAIN does not resolve to any IP."
        echo "Create an A record: $DOMAIN -> $PUBLIC_IP"
        exit 1
    elif [ "$RESOLVED_IP" != "$PUBLIC_IP" ]; then
        echo "WARNING: $DOMAIN resolves to $RESOLVED_IP but this VM's IP is $PUBLIC_IP"
        echo "DNS may not have propagated yet, or the record points elsewhere."
        echo "Continuing anyway — Caddy will fail to get TLS if DNS is wrong."
    else
        echo "OK: $DOMAIN -> $RESOLVED_IP (matches this VM)"
    fi
fi

# 2. Check ports
echo ""
echo "[2/4] Checking that ports 80 and 443 are free..."
for port in 80 443; do
    if ss -tlnp | grep -q ":${port} "; then
        echo "WARNING: Port $port is already in use. Caddy may fail to bind."
        echo "         Stop the service using it, or change Caddy's ports."
    else
        echo "OK: Port $port is free."
    fi
done

# 3. Verify .env exists
echo ""
echo "[3/4] Checking .env..."
if [ ! -f "$REPO_DIR/.env" ]; then
    echo "Creating .env from .env.example..."
    cp "$REPO_DIR/.env.example" "$REPO_DIR/.env"
    echo "IMPORTANT: Edit .env before proceeding! Set at minimum:"
    echo "  - JWT_SECRET"
    echo "  - BACKEND_ORIGIN=https://$DOMAIN"
    echo "  - RESEND_API_KEY"
    echo ""
    echo "Run this script again after editing .env."
    exit 0
else
    echo "OK: .env exists."
    # Quick sanity check
    if grep -q 'change-me-in-production\|CHANGE_ME' "$REPO_DIR/.env" 2>/dev/null; then
        echo "WARNING: .env still contains default/placeholder values."
        echo "         Make sure JWT_SECRET and other secrets are changed."
    fi
fi

# 4. Start the stack
echo ""
echo "[4/4] Starting Docker Compose..."
cd "$REPO_DIR"
docker compose -f docker-compose.selfhost.yml up -d

echo ""
echo "=== Setup Complete ==="
echo ""
echo "Caddy will automatically provision TLS for $DOMAIN via Let's Encrypt."
echo "This may take 1-2 minutes on first start."
echo ""
echo "Check status:    docker compose -f docker-compose.selfhost.yml ps"
echo "Check Caddy logs: docker compose -f docker-compose.selfhost.yml logs caddy"
echo "Check API health: curl -sf https://$DOMAIN/health"
echo ""
echo "Resend webhook:   https://$DOMAIN/api/webhooks/resend"
echo "Tracking pixel:   https://$DOMAIN/track/pixel/{email_log_id}"
echo "Tracking click:   https://$DOMAIN/track/click/{email_log_id}?url={base64_url}"