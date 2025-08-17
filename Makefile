VERSION_PLUGIN := $(shell cat ./components/manawell-device-plugin/.Version)
VERSION_ENCHANTER := $(shell cat ./components/runesmith-enchanter/.Version)
VERSION_BACKEND := $(shell cat ./components/runesmith-backend/.Version)
VERSION_DASHBOARD := $(shell cat ./components/runesmith-dashboard/.Version)


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

#backend commands
backend-run:
	cd ./components/runesmith-backend && go run ./cmd/server/main.go --config=config.example.yaml
backend-load-config:
	cd ./components/runesmith-backend && go run ./cmd/server/main.go load-config --config=config.example.yaml
backend-helm-template:
	helm template deployment/helm/runesmith-backend
backend-docker-build:
	docker build --no-cache --debug -f components/runesmith-backend/Dockerfile --build-arg FULL_VERSION=$(VERSION_BACKEND).0 -t runesmith-backend:latest .

#dashboard commands
dashboard-run:
	cd ./components/runesmith-dashboard && npm run dev
dashboard-helm-template:
	helm template deployment/helm/runesmith-dashboard
dashboard-docker-build:
	docker build --no-cache --debug -f components/runesmith-dashboard/Dockerfile --build-arg FULL_VERSION=$(VERSION_DASHBOARD).0 -t runesmith-dashboard:latest .