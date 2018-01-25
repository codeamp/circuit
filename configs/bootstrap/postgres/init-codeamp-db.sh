#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
	CREATE USER codeamp;
  CREATE DATABASE codeamp;
	GRANT ALL PRIVILEGES ON DATABASE codeamp TO codeamp;
  \c codeamp
  CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
  CREATE EXTENSION IF NOT EXISTS hstore;
EOSQL
