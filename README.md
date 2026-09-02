# ACCEL-PPP Miniadmin

## I dont want to read, just give me the commands

Debian 13 (trixie)
```bash
apt update
apt install git lsof docker.io docker-compose
git clone https://github.com/nuclearcat/accel-ppp-miniadmin.git
cd accel-ppp-miniadmin/install
sh install.sh
```

Ubuntu 24.04 LTS and Ubuntu 26.04 LTS
```bash
apt update
apt install git lsof docker.io docker-compose-v2
git clone https://github.com/nuclearcat/accel-ppp-miniadmin.git
cd accel-ppp-miniadmin/install
sh install.sh
```

Note the different compose package per distribution. On Debian 13 the
`docker-compose` package is already Compose v2. On Ubuntu it is not: 24.04 still
ships the end-of-life Python Compose v1 under that name, and 26.04 dropped the
name entirely, so on both you want `docker-compose-v2`. `install.sh` works with
either v1 or v2. Debian 12 also still works, using the same command as Debian 13.

After that it will ask you two questions:
- Please enter secret token for SSTP Admin interface

This is password for web interface

- Please enter SSTP server hostname

You need real domain name pointing to this server. Some hosting providers might provide you with subdomain, like **x-x-x-x.ip.linodeusercontent.com**

It cannot be an IP address - this name is what Let's Encrypt issues the certificate for, and certbot cannot validate an IP. `install.sh` rejects one, and warns if the name does not resolve yet.

After that you can access web interface at https://yourhostname:8080 , login, add user/password, then just spin up SSTP client (available by default in Windows) and connect to your server.

## Description

This is simple web interface for ACCEL-PPP based VPN. It is mostly created to operate with accel-ppp-docker repository, to create easy to use VPN install, but in future might work with other setups as well.
**At current stage it is very basic and only allows to add/remove users and related docker image supports only SSTP VPN.**

## Features

- Token based authentication
- Add/Remove users web interface and API

## Installation

You need host with ports 443(for SSTP), 80(for Let's encrypt certbot) and 8080(for miniadmin) available and open. Also you need hostname pointing to this host.

Required packages: `git`, `lsof`, docker and docker compose - see the
distribution specific commands above. `lsof` is used by `install.sh` to check
that the required ports are actually free before starting anything.

```bash
git clone https://github.com/nuclearcat/accel-ppp-miniadmin.git
cd accel-ppp-miniadmin/install
sh install.sh
```

## Configuration

.env file contains all configuration options - hostname and token. If you want to update token, just change it in .env file and restart daemons using `docker-compose restart` (or `docker compose restart` on Compose v2).

## Usage

After installation, you can access miniadmin web interface at http://yourhostname:8080. 

## API

API is available at http://yourhostname:8080/api. You can use it to add/remove users. API is token based, so you need to provide token in header.

### Authentication

Set token in header:

```txt
Authorization: Bearer yourtoken
```

### Add user

POST /api

```json
{
    "action": "add",
    "username": "test",
    "password": "test"
}
```

### Remove user

POST /api

```json
{
    "action": "delete",
    "username": "test"
}
```

### Show users

POST /api

```json

{
    "action": "show"
}
```
