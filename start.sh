#!/usr/bin/env bash
export MYSQL_ADDR=192.168.1.119
export MYSQL_PORT=3306
export MYSQL_DATABASE=mysql
export MYSQL_USER=root
export MYSQL_ENV_MYSQL_ROOT_PASSWORD=root

go build
./broker_mysql