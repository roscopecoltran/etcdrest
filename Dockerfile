FROM gliderlabs/alpine

COPY bin/etcdrest /etcdrest

RUN mkdir /schemas /templates

COPY templates/print.tmpl /templates/print.tmpl

EXPOSE 8080
ENTRYPOINT ["/etcdrest"]
