NAME:=etcd-rest
#RELEASE:=$(shell git rev-parse --verify --short HEAD)
#SITE:=$(shell git config --get remote.origin.url | sed -e 's/^.*@//' -e 's/:.*$$//')
#PKG:=$(shell git config --get remote.origin.url | sed -e 's/^.*://' -e 's/\.git$$//')
#USER=mickep76
SRCDIR=src/github.com/mickep76
TMPDIR1=.build
VERSION:=$(shell awk -F '"' '/Version/ {print $$2}' ${SRCDIR}/common/version.go)
RELEASE:=$(shell date -u +%Y%m%d%H%M)
ARCH:=$(shell uname -p)

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

build: test
        gb build all

install:
	cp bin/* /usr/local/bin

rpm: build
	mkdir -p ${TMPDIR1}/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS}
	cp -r bin ${TMPDIR1}/SOURCES
	sed -e "s/%NAME%/${NAME}/g" -e "s/%VERSION%/${VERSION}/g" -e "s/%RELEASE%/${RELEASE}/g" \
		${NAME}.spec >${TMPDIR1}/SPECS/${NAME}.spec
	rpmbuild -vv -bb --target="${ARCH}" --clean --define "_topdir $$(pwd)/${TMPDIR1}" ${TMPDIR1}/SPECS/${NAME}.spec
	mv ${TMPDIR1}/RPMS/${ARCH}/*.rpm .

#docker-clean:
#	docker rmi ${NAME} &>/dev/null || true

#docker-build: docker-clean
#	docker run --rm -it -v "$$PWD":/go -w /go mickep76/alpine-golang:latest
#	docker build --pull=true --no-cache -t ${USER}/${NAME}:${RELEASE} .
#	docker tag -f ${USER}/${NAME}:${RELEASE} ${USER}/${NAME}:latest

#docker-push: docker-build
#	docker login -u ${USER}
#	docker push ${USER}/${NAME}:${RELEASE}
#	docker push ${USER}/${NAME}:latest
