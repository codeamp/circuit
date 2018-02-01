#!/bin/sh

#Wait for postgres to start
while ! curl http://postgres:$DOCKER_COMPOSE_POSTGRES_PORT/ 2>&1 | grep '52'
do
  echo "Waiting for postgres to start..sleeping for 1 second"
  sleep 1
done

/usr/local/bin/dex serve configs/dex.yml & reflex -c reflex.conf