# Circuit [![CircleCI](https://circleci.com/gh/codeamp/circuit.svg?style=svg)](https://circleci.com/gh/codeamp/circuit) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![codecov](https://codecov.io/gh/codeamp/circuit/branch/master/graph/badge.svg)](https://codecov.io/gh/codeamp/circuit) [![Go Report Card](https://goreportcard.com/badge/codeamp/circuit)](https://goreportcard.com/report/codeamp/circuit) [![codebeat badge](https://codebeat.co/badges/b977a7e7-1e94-43e1-9e58-463cff99add3)](https://codebeat.co/projects/github-com-codeamp-circuit-master)
This is the API layer of the overall Codeamp project. It is built with Golang, GraphQL, GORM and Socket-IO.

## Installation

1. `git clone https://github.com/codeamp/circuit.git`
2. `cp configs/circuit.yml configs/circuit.dev.yml`
3. `make up`
4. Go to `localhost:3011` once you see events being processed in your command line. If all is well, you should see a GraphiQL client.

## Dev with skaffold

1. Run `skaffold run`

2. Init circuit db `cat .gitops/db-backups/circuit-dev-backup.tar | kubectl exec -n codeamp -i postgres-0 -- pg_restore --dbname=postgresql://postgres:password@127.0.0.1:5432/codeamp -v`

3. Need to port-forward for dex because browser needs to talk to same host as circuit app ````kubectl port-forward `kubectl get po -l app=circuit -o name` 5556:5556```
