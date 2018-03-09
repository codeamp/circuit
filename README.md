# Circuit

This is the API layer of the overall Codeamp project. It is built with Golang, GraphQL, GORM and Socket-IO.


## Installation

1. `git clone https://github.com/codeamp/circuit.git`
2. `cp configs/circuit.yml configs/circuit.dev.yml`
3. `make build && make up`


## Testing

### Resolvers
1. Create a db called `codeamp_test`
2. `cd plugins/codeamp/schema/resolvers/`
3. `go test -v`
