codegen:
	oapi-codegen --config api/gen/config.yaml spec/swagger.yaml

devUp:
	docker compose -f docker/docker-compose.yml up -d

devDown:
	docker compose -f docker/docker-compose.yml down

build:
	go build .

build-docker:
	docker build -f docker/Dockerfile -t flynn/shoulder .

