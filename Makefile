NAME:=$(shell basename `git rev-parse --show-toplevel`)
RELEASE:=$(shell git rev-parse --verify --short HEAD)
SITE:=$(shell git config --get remote.origin.url | sed -e 's/^.*@//' -e 's/:.*$$//')
PKG:=$(shell git config --get remote.origin.url | sed -e 's/^.*://' -e 's/\.git$$//')
USER=mickep76

all: build

clean:
	rm -rf pkg bin

test: clean
	gofmt -w=true src/${SITE}/${PKG}
	golint src/${SITE}/${PKG}
	GOPATH=$$PWD; go vet ${SITE}/${PKG}
	gb test

build: test
	gb build all

update:
	gb vendor update --all

docker-clean:
	docker rmi ${NAME} &>/dev/null || true

docker-build: docker-clean
	docker run --rm -it -v "$$PWD":/go -w /go mickep76/alpine-golang:latest
	docker build --pull=true --no-cache -t ${USER}/${NAME}:${RELEASE} .
	docker tag -f ${USER}/${NAME}:${RELEASE} ${USER}/${NAME}:latest

docker-push: docker-build
	docker login -u ${USER}
	docker push ${USER}/${NAME}:${RELEASE}
	docker push ${USER}/${NAME}:latest
