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

if ! [ -x "$(command -v docker-compose)" ]; then
    echo "Error: docker-compose is not installed."
    exit 1
fi

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
docker-compose up -d

echo "Web interface available at https://${SSTP_HOSTNAME}:8080"


