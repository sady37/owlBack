#!/usr/bin/env bash
# =============================================================================
# install-docker.sh
# Docker Engine installation script for Ubuntu (22.04 / 24.04, x86_64 / arm64)
#
# What this script does:
#   1. Installs Docker Engine, CLI, containerd, buildx, and compose plugin
#   2. Configures registry mirrors for faster pulls in mainland China
#   3. Enables and starts the Docker service
#   4. Adds the current user to the "docker" group
#   5. Verifies the installation by running the hello-world container
#
# Usage:
#   chmod +x install-docker.sh
#   ./install-docker.sh
#
# Requirements:
#   - Ubuntu 22.04 LTS (Jammy) or 24.04 LTS (Noble)
#   - x86_64 or arm64 architecture
#   - sudo privileges (passwordless sudo is recommended)
# =============================================================================

set -euo pipefail

# ---------------------------------------------------------------------------
# Helper functions
# ---------------------------------------------------------------------------
log()  { echo "[INFO]  $*"; }
warn() { echo "[WARN]  $*" >&2; }
die()  { echo "[ERROR] $*" >&2; exit 1; }

# ---------------------------------------------------------------------------
# 1. Verify the OS is a supported Ubuntu release
# ---------------------------------------------------------------------------
if ! grep -qi ubuntu /etc/os-release 2>/dev/null; then
    die "This script only supports Ubuntu. Detected OS: $(. /etc/os-release && echo "$PRETTY_NAME")"
fi

UBUNTU_CODENAME=$(. /etc/os-release && echo "$VERSION_CODENAME")
ARCH=$(dpkg --print-architecture)

log "Detected: Ubuntu ${UBUNTU_CODENAME} (${ARCH})"

if [[ "$ARCH" != "amd64" && "$ARCH" != "arm64" ]]; then
    die "Unsupported architecture: $ARCH. Only amd64 and arm64 are supported."
fi

# ---------------------------------------------------------------------------
# 2. Remove any conflicting legacy Docker packages
# ---------------------------------------------------------------------------
log "Removing conflicting legacy Docker packages (if any)..."
for pkg in docker.io docker-doc docker-compose docker-compose-v2 podman-docker containerd runc; do
    sudo apt-get remove -y "$pkg" 2>/dev/null || true
done

# ---------------------------------------------------------------------------
# 3. Install prerequisites
# ---------------------------------------------------------------------------
log "Updating package index and installing prerequisites..."
sudo apt-get update -qq
sudo apt-get install -y ca-certificates curl gnupg lsb-release

# ---------------------------------------------------------------------------
# 4. Add Docker's official GPG key and apt repository
# ---------------------------------------------------------------------------
log "Adding Docker GPG key..."
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL "https://download.docker.com/linux/ubuntu/gpg" \
    | sudo gpg --dearmor --batch --yes -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

log "Adding Docker apt repository..."
echo "deb [arch=${ARCH} signed-by=/etc/apt/keyrings/docker.gpg] \
https://download.docker.com/linux/ubuntu ${UBUNTU_CODENAME} stable" \
    | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# ---------------------------------------------------------------------------
# 5. Install Docker Engine and plugins
# ---------------------------------------------------------------------------
log "Installing Docker Engine, CLI, containerd, buildx, and compose plugin..."
sudo apt-get update -qq
sudo apt-get install -y \
    docker-ce \
    docker-ce-cli \
    containerd.io \
    docker-buildx-plugin \
    docker-compose-plugin

# ---------------------------------------------------------------------------
# 6. Configure registry mirrors (optimised for mainland China)
# ---------------------------------------------------------------------------
log "Configuring registry mirrors..."
sudo mkdir -p /etc/docker
sudo tee /etc/docker/daemon.json > /dev/null <<'DAEMON_JSON'
{
  "registry-mirrors": [
    "https://docker.1ms.run",
    "https://docker.xuanyuan.me"
  ],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m",
    "max-file": "3"
  }
}
DAEMON_JSON

# ---------------------------------------------------------------------------
# 7. Enable and start the Docker service
# ---------------------------------------------------------------------------
log "Enabling and starting Docker service..."
sudo systemctl enable docker --now
sudo systemctl is-active docker || die "Docker service failed to start."

# ---------------------------------------------------------------------------
# 8. Add the current user to the "docker" group
# ---------------------------------------------------------------------------
CURRENT_USER="${SUDO_USER:-$USER}"
if ! id -nG "$CURRENT_USER" | grep -qw docker; then
    log "Adding user '${CURRENT_USER}' to the 'docker' group..."
    sudo usermod -aG docker "$CURRENT_USER"
    warn "Group change takes effect after you log out and back in."
else
    log "User '${CURRENT_USER}' is already in the 'docker' group."
fi

# ---------------------------------------------------------------------------
# 9. Verify installation with hello-world container
# ---------------------------------------------------------------------------
log "Running hello-world container to verify Docker..."
sudo docker run --rm hello-world

# ---------------------------------------------------------------------------
# 10. Print version summary
# ---------------------------------------------------------------------------
log "------------------------------------------------------------"
log "Docker installation complete!"
sudo docker version --format "  Engine version : {{.Server.Version}}"
sudo docker compose version --short | xargs -I{} echo "  Compose version: {}"
log "------------------------------------------------------------"
log "NOTE: Log out and log back in (or run 'newgrp docker') to use"
log "      docker without sudo."
