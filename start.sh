#!/usr/bin/env bash
export MYSQL_ADDR=54.223.94.93
export MYSQL_PORT=3306
export MYSQL_DATABASE=mysql
export MYSQL_USER=root
export MYSQL_ENV_MYSQL_ROOT_PASSWORD=root

go build
./broker_mysql