FROM gliderlabs/alpine

RUN mkdir -p /app/bin /app/etc /app/schemas /app/templates /app/logs

COPY bin/etcdrest /app/bin/etcdrest

COPY templates/print.tmpl /app/templates/print.tmpl

# Fix name resolution including /etc/hosts
RUN echo 'hosts: files mdns4_minimal [NOTFOUND=return] dns mdns4' >> /etc/nsswitch.conf

EXPOSE 8080
ENTRYPOINT ["/app/bin/etcdrest"]
