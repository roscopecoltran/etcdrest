FROM gliderlabs/alpine

COPY bin/etcdrest /etcdrest

EXPOSE 8080
ENTRYPOINT ["/etcdrest"]
