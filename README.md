# Circuit

This is the API layer of the overall Codeamp project. It is built with Golang, GraphQL, GORM and Socket-IO.


## Installation

1. `git clone https://github.com/codeamp/circuit.git`
2. `cp configs/circuit.yml configs/circuit.dev.yml`
3. `docker-compose up redis postgres circuit`
4. `go run main.go migrate --config configs/circuit.dev.yml`
5. `go run main.go start --config configs/circuit.dev.yml`


## Testing

### Resolvers
1. Create a db called `codeamp_test`
2. `cd plugins/codeamp/schema/resolvers/`
3. `go test -v`

**Current Tests**
- [X] Project 
- [X] Feature
- [ ] Environment
- [ ] Environment Variable
- [ ] Extension
- [ ] Extension Spec
- [ ] Release Extension
- [ ] Release
- [ ] Service Spec
- [ ] User
- [ ] Service

## TODO

- [ ] Install default extensions depending on project type
- [ ] Create a full separation of environments where project features, releases, extensions and settings will be different depending on the environment context the user is in.
- [ ]  Implement CI and any other standard extensions that will be almost certainly be required for project deployments.
- [ ]  Implement Access Control to restrict views based on permissioning, either with an external API (e.g. Dex, Okta) or in-house.
- [ ]  Implement Audit Trail to view all actions users have made in a particular project (possibly in the form of an extension)
- [ ] Implement a dashboard with relevant graphs and metrics for admins and user-specific deployments.
- [ ] Implement relevant views and metrics for the environments manager.
- [ ] Map branches to specific environments (either on a project or global basis)
- [ ] Documentation on creating an Extension and guided walkthrough
- [ ] Writing backend unit and integration tests
