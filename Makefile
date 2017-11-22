up:
	docker-compose up -d redis postgres
	docker-compose run --rm circuit go run main.go --config ./configs/circuit.dev.yml migrate
	docker-compose up circuit

build:
	docker-compose build --pull circuit

destroy:
	docker-compose stop
	docker-compose rm -f

.PHONY: up build destroy # let's go to reserve rules names
