#!/usr/bin/env bash
export MYSQL_ADDR=127.0.0.1
export MYSQL_PORT=3306
export MYSQL_DATABASE=mysql
export MYSQL_USER=root
export MYSQL_ENV_MYSQL_ROOT_PASSWORD=root

go build
./datafactory-servicebroker-mysql
