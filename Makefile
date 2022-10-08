debugUp:
	docker compose -f docker/docker-compose-debug.yml up -d 

allUp:
	docker compose -f docker/docker-compose.yml up -d

allDown:
	docker compose -f docker/docker-compose.yml down --remove-orphans

debugDown:
	docker compose -f docker/docker-compose-debug.yml down --remove-orphans

build-docker:
	oapi-codegen --config api/gen/config.yaml spec/swagger.yaml
	docker build -f docker/Dockerfile -t flynn/shoulder .

k8s-deploy:
	eval $(minikube docker-env)
	oapi-codegen --config api/gen/config.yaml spec/swagger.yaml
	docker build -f docker/Dockerfile -t flynn/shoulder .

	docker pull postgres
	docker pull rabbitmq

	minikube kubectl -- apply -f deploy/shoulder.yaml

	minikube kubectl -- expose deployment rabbitmq --target-port=5672
	minikube kubectl -- expose deployment postgres --target-port=5432
	minikube kubectl -- expose deployment shoulder --port=8080
