# Circuit
[![CircleCI](https://circleci.com/gh/codeamp/circuit.svg?style=svg)](https://circleci.com/gh/codeamp/circuit)

This is the API layer of the overall Codeamp project. It is built with Golang, GraphQL, GORM and Socket-IO.


## Installation

1. `git clone https://github.com/codeamp/circuit.git`
2. `cp configs/circuit.yml configs/circuit.dev.yml`
3. `make up`
4. Go to `localhost:3011` once you see events being processed in your command line. If all is well, you should see a GraphiQL client.
