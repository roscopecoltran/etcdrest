NAME=etcdrest
VERSION=0.3
USER=mickep76
BUILD=.build

all: build

clean:
	rm -rf pkg bin ${BUILD}

build:
	gb build all

docker-clean: clean
	docker rmi ${NAME} &>/dev/null || true

docker: docker-main docker-example

docker-main: docker-clean
	docker pull mickep76/alpine-golang:latest
	docker run --rm -it -v "$$PWD":/app -w /app mickep76/alpine-golang:latest
	docker build --pull=true --no-cache -t ${USER}/${NAME}:${VERSION} .
	docker tag -f ${USER}/${NAME}:${VERSION} ${USER}/${NAME}:latest

docker-example:
	( cd example; docker build --pull=true --no-cache -t ${USER}/${NAME}-example:${VERSION} . )
	docker tag -f ${USER}/${NAME}-example:${VERSION} ${USER}/${NAME}-example:latest

push: docker push-main push-example

push-main:
	docker login -u ${USER}
	docker push ${USER}/${NAME}:${VERSION}
	docker push ${USER}/${NAME}:latest

push-example:
	docker push ${USER}/${NAME}-example:latest
