devUp:
	docker compose -f docker/docker-compose.yml up -d

devDown:
	docker compose -f docker/docker-compose.yml down

build:
	go build .

build-docker:
	oapi-codegen --config api/gen/config.yaml spec/swagger.yaml
	docker build -f docker/Dockerfile -t flynn/shoulder .

k8s-deploy:
	eval $(minikube -p minikube docker-env)

	kubectl apply -f deploy/shoulder.yaml

	kubectl expose deployment rabbitmq --target-port=5672
	kubectl expose deployment postgres --target-port=5432
	kubectl expose deployment shoulder --port=8080
