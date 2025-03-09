VERSION = $(shell git describe --tags | cut -c 2-)

.PHONY: docker-build
docker-build:
	docker build --build-arg APP_VERSION=${VERSION} -t jabbors/euribor:${VERSION} .

.PHONY: docker-release
docker-release: docker-build
	docker push jabbors/euribor:${VERSION}
