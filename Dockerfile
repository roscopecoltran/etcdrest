FROM gliderlabs/alpine

COPY bin/etcd-rest /etcd-rest

EXPOSE 3000
CMD ["/etcd-rest"]
