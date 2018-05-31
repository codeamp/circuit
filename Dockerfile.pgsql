FROM postgres

COPY ./bootstrap/postgres/docker-entrypoint-initdb.d /docker-entrypoint-initdb.d
COPY ./bootstrap/postgres/data /var/lib/postgresql/data