NAME=etcdrest
VERSION=0.3
USER=mickep76
BUILD=.build

all: build

docker: docker-build

push: docker-push

clean:
	rm -rf pkg bin ${BUILD}

build:
	gb build all

docker-clean: clean
	docker rmi ${NAME} &>/dev/null || true

docker-build: docker-clean
	docker pull mickep76/alpine-golang:latest
	docker run --rm -it -v "$$PWD":/app -w /app mickep76/alpine-golang:latest
	docker build --pull=true --no-cache -t ${USER}/${NAME}:${VERSION} .

docker-push: docker-build
	docker login -u ${USER}
	docker push ${USER}/${NAME}:${VERSION}
	docker tag -f ${USER}/${NAME}:${VERSION} ${USER}/${NAME}:latest
	docker push ${USER}/${NAME}:latest
