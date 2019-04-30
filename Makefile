SERVICE="circuit"
CONFIG="./configs/circuit.yml"

up:
	docker-compose up -d redis postgres mysql
	docker-compose run --rm ${SERVICE} go run main.go migrate --config ${CONFIG}	
	docker-compose up ${SERVICE}

build:
	docker-compose build --pull ${SERVICE}

destroy:
	docker-compose stop
	docker-compose rm -f

assets:
	docker-compose run --rm ${SERVICE} go-bindata -pkg assets -o assets/assets.go \
		plugins/codeamp/graphql/schema.graphql \
		plugins/codeamp/graphql/static/

.PHONY: up build destroy assets
