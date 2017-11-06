#!/bin/bash -e

_term() { 
  echo "Caught SIGTERM signal!" 
  while :
  do
    if [[ $(ps aux |grep -v grep | grep tmux | wc -l) = "0" ]]; then
      echo "No running TMUX sessions! Bye Bye!" 
      kill -TERM "$child" 2>/dev/null
    fi

    echo "Found running TMUX sessions! Sleeping..." 
    sleep 60
  done
}

trap _term SIGTERM

mkdir -p /opt ~/.tsh/
cd /opt

apt-get update && apt-get install -y tmux

curl -L -o teleport. https://teleport-static.checkrhq.net/teleport.tar.gz && \
  tar -xzvf teleport.tar.gz && \
  rm -f teleport.tar.gz && \
  mv teleport-ent/teleport teleport-ent/tsh teleport-ent/tctl /usr/bin

mkdir -p /var/lib/teleport/log /root/.tsh/

env > /root/.tsh/environment

teleport start --permit-user-env --roles=node --token=$TELEPORT_TOKEN_V2 --auth-server=$TELEPORT_AUTH_SERVER_V2 --nodename=$CODEFLOW_SLUG::$CODEFLOW_CREATED_AT --labels=created_at=$CODEFLOW_CREATED_AT,hash=$CODEFLOW_HASH

child=$!
