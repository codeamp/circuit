up:
	docker-compose run --rm circuit  go run main.go --config ./configs/circuit.dev.yml migrate up
	docker-compose up redis postgres circuit

build:
	docker-compose build circuit

destroy:
	docker-compose stop
	docker-compose rm -f

.PHONY: up build destroy # let's go to reserve rules names

