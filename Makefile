up:
	docker-compose up -d redis postgres
	docker-compose run --rm circuit go run main.go migrate --config ./configs/circuit.yml	
	docker-compose up circuit

build:
	docker-compose build --pull circuit

destroy:
	docker-compose stop
	docker-compose rm -f

.PHONY: up build destroy # let's go to reserve rules names
