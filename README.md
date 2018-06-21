# Circuit [![CircleCI](https://circleci.com/gh/codeamp/circuit.svg?style=svg)](https://circleci.com/gh/codeamp/circuit) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![Coverage Status](https://coveralls.io/repos/github/codeamp/circuit/badge.svg?branch=master)](https://coveralls.io/github/codeamp/circuit?branch=master) [![Go Report Card](https://goreportcard.com/badge/codeamp/circuit)](https://goreportcard.com/report/codeamp/circuit) [![codebeat badge](https://codebeat.co/badges/b977a7e7-1e94-43e1-9e58-463cff99add3)](https://codebeat.co/projects/github-com-codeamp-circuit-master)
This is the API layer of the overall Codeamp project. It is built with Golang, GraphQL, GORM and Socket-IO.

## Installation

1. `git clone https://github.com/codeamp/circuit.git`
2. `cp configs/circuit.yml configs/circuit.dev.yml`
3. `make up`
4. Go to `localhost:3011` once you see events being processed in your command line. If all is well, you should see a GraphiQL client.
