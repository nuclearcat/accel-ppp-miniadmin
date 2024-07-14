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

echo "Please enter secret token for SSTP Admin interface"
read -s SSTP_ADMINTOKEN
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

docker-compose up -d



