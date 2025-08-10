gen-proto:
	protoc --go_out=. --go-grpc_out=. manawell.proto

helm-template:
	helm template ./../../deployment/helm/manawell-device-plugin

build-docker-test:
	docker build --no-cache -t manawell-device-plugin:dev .