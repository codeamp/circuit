#!/bin/bash
set -e

export PGPASSWORD=$POSTGRES_PASSWORD
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
	CREATE USER codeamp;
  CREATE DATABASE codeamp;
  CREATE DATABASE codeamp_test;

  GRANT ALL PRIVILEGES ON DATABASE codeamp TO codeamp;
  GRANT ALL PRIVILEGES ON DATABASE codeamp_test TO codeamp;

  \c codeamp
  CREATE EXTENSION IF NOT EXISTS pgcrypto;
  CREATE EXTENSION IF NOT EXISTS hstore;
  
  \c codeamp_test
  CREATE EXTENSION IF NOT EXISTS pgcrypto;
  CREATE EXTENSION IF NOT EXISTS hstore; 
EOSQL
