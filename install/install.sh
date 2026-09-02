#!/bin/sh
# if this is just standalone install.sh we will not find docker-compose.yaml
if [ ! -e docker-compose.yaml ]; then
    # is git installed?
    if ! [ -x "$(command -v git)" ]; then
        echo "Error: git is not installed."
        apt-get install git
    fi
    git clone https://github.com/nuclearcat/accel-ppp-miniadmin
    cd accel-ppp-miniadmin/install
fi

# is docker installed? and docker-compose?
if ! [ -x "$(command -v docker)" ]; then
    echo "Error: docker is not installed."
    exit 1
fi

# Probe compose by running it, not by looking for a binary: the v2 plugin is
# invoked as "docker compose" and has no executable of its own to find.
# Prefer v2, the standalone "docker-compose" may still be the end-of-life v1.
if docker compose version >/dev/null 2>&1; then
    DOCKER_COMPOSE="docker compose"
elif docker-compose version >/dev/null 2>&1; then
    DOCKER_COMPOSE="docker-compose"
else
    echo "Error: neither docker compose (v2) nor docker-compose (v1) is available."
    echo "Please install the compose package for your distribution:"
    echo "  Debian 13:                apt install docker-compose"
    echo "  Ubuntu 24.04 and 26.04:   apt install docker-compose-v2"
    exit 1
fi

echo "Using: ${DOCKER_COMPOSE} ($(${DOCKER_COMPOSE} version 2>/dev/null | head -1))"

# verify if port 80,443,8080 available
if [ -n "$(lsof -i :80)" ]; then
    echo "Port 80 is already in use"
    exit 1
fi
if [ -n "$(lsof -i :443)" ]; then
    echo "Port 443 is already in use"
    exit 1
fi
if [ -n "$(lsof -i :8080)" ]; then
    echo "Port 8080 is already in use"
    exit 1
fi

echo "Please enter secret token for SSTP Admin interface"
read SSTP_ADMINTOKEN
echo "Please enter SSTP server hostname"
read SSTP_HOSTNAME

# Verify SSTP_ADMINTOKEN
if [ -z "${SSTP_ADMINTOKEN}" ]; then
    echo "SSTP_ADMINTOKEN is not set"
    exit 1
fi

# Verify SSTP_HOSTNAME
if [ -z "${SSTP_HOSTNAME}" ]; then
    echo "SSTP_HOSTNAME is not set"
    exit 1
fi

echo "SSTP_ADMINTOKEN=${SSTP_ADMINTOKEN}" > .env
echo "SSTP_HOSTNAME=${SSTP_HOSTNAME}" >> .env

mkdir -p accel-letsencrypt accel-ppp

echo "Waiting for accel-ppp to start"
$DOCKER_COMPOSE up -d

echo "Web interface available at https://${SSTP_HOSTNAME}:8080"


