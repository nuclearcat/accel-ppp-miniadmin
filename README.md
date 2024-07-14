# ACCEL-PPP Miniadmin

## Description

This is simple web interface for ACCEL-PPP based VPN. It is mostly created to operate with accel-ppp-docker repository, to create easy to use VPN install, but in future might work with other setups as well.

## Features

- Token based authentication
- Add/Remove users web interface and API

## Installation

You need host with ports 443(for SSTP), 80(for Let's encrypt certbot) and 8080(for miniadmin) available and open. Also you need hostname pointing to this host.

```bash
git clone git@github.com:nuclearcat/accel-ppp-miniadmin.git
cd accel-ppp-miniadmin/install
sh install.sh
```

## Configuration

.env file contains all configuration options - hostname and token. If you want to update token, just change it in .env file and restart docker-compose using docker-compose restart.

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
