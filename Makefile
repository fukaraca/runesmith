VERSION_PLUGIN := $(shell cat ./components/manawell-device-plugin/.Version)
VERSION_ENCHANTER := $(shell cat ./components/runesmith-enchanter/.Version)


# Plugin commands
plugin-load-config:
	cd ./components/manawell-device-plugin && go run . load-config
plugin-helm-template:
	helm template deployment/helm/manawell-device-plugin
plugin-docker-build:
	docker build --no-cache --debug -f components/manawell-device-plugin/Dockerfile --build-arg FULL_VERSION=$(VERSION_PLUGIN).0 -t manawell-device-plugin:latest .

# enchanter commands
enchanter-run:
	@echo "did you set env variables"
	go run ./components/runesmith-enchanter/
enchanter-docker-build:
	docker build --no-cache --debug -f components/runesmith-enchanter/Dockerfile --build-arg FULL_VERSION=$(VERSION_ENCHANTER).0 -t runesmith-enchanter:latest .