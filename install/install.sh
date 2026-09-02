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

# The client alone is not enough: a half-configured package (for example an
# aborted dpkg run) leaves /usr/bin/docker in place with no daemon behind it,
# and every command below would fail only at the very end.
if ! docker info >/dev/null 2>&1; then
    echo "Error: cannot talk to the docker daemon at unix:///var/run/docker.sock."
    echo "Check that it is installed and running, and that you have access to it:"
    echo "  systemctl enable --now docker"
    echo "  docker info"
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
# lsof must be present: without it the command substitution below is empty and
# every port would look free, so the install would happily proceed onto ports
# that are already taken.
if ! [ -x "$(command -v lsof)" ]; then
    echo "Error: lsof is not installed, cannot verify that ports 80, 443 and 8080 are free."
    echo "Please install it: apt install lsof"
    exit 1
fi

# lsof only reports sockets of processes we may inspect, so a non-root run can
# miss a listener owned by someone else and wrongly report the port as free.
if [ "$(id -u)" -ne 0 ]; then
    echo "Warning: not running as root, the port checks below may miss listeners owned by other users."
fi

for PORT in 80 443 8080; do
    if [ -n "$(lsof -i :${PORT})" ]; then
        echo "Port ${PORT} is already in use"
        exit 1
    fi
done

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
# Do not announce success on a failed start: without this check a pull error or
# an unreachable daemon still printed the web interface URL below.
if ! $DOCKER_COMPOSE up -d; then
    echo "Error: ${DOCKER_COMPOSE} up failed, the containers are not running."
    echo "Fix the error above and re-run: ${DOCKER_COMPOSE} up -d"
    exit 1
fi

echo "Web interface available at https://${SSTP_HOSTNAME}:8080"


