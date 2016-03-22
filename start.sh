#!/usr/bin/env bash
export MYSQL_PORT_3306_TCP_ADDR=10.1.236.121
export MYSQL_PORT_3306_TCP_PORT=3306
export MYSQL_DATABASE=mysql
export MYSQL_USER=root
export MYSQL_ENV_MYSQL_ROOT_PASSWORD=root

go build
./datafactory-servicebroker-mysql