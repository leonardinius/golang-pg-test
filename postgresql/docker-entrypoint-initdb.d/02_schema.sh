#!/bin/bash

set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" dbuser <<-EOSQL
  CREATE SCHEMA IF NOT EXISTS "dbuser" authorization "dbuser";
EOSQL

