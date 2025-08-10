VERSION_PLUGIN := $(shell cat ./components/manawell-device-plugin/.Version)


# Plugin commands
plugin-load-config:
	cd ./components/manawell-device-plugin && go run . load-config
plugin-helm-template:
	helm template deployment/helm/manawell-device-plugin
plugin-docker-build:
	docker build -f components/manawell-device-plugin/Dockerfile --build-arg FULL_VERSION=$(VERSION_PLUGIN).0 -t manawell-device-plugin:latest .